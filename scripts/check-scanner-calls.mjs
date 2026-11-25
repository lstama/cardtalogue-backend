#!/usr/bin/env node
// Owner-side "no orphan calls" check for the backend → scanner boundary.
//
// The contract repo no longer enforces this (it cannot see the consumers); each
// consumer checks its own calls against the contract version it pins. This
// script is dependency-free (Node stdlib only) and run by
// .github/workflows/contract-check.yml.
//
// What it does
//   1. Resolves the scanner URL the backend is configured with
//      (`CARD_SCAN_API_URL`: process env > .env > internal/config/config.go
//      default > Makefile).
//   2. Scans non-test Go sources for outbound HTTP calls
//      (`http.NewRequest` / `http.NewRequestWithContext`) that target the
//      scanner, tracing the URL variable through the constructor to the config
//      field, plus any scanner URL literal (host:port match).
//   3. Asserts every scanned path (and method, when it can be resolved) is
//      defined in the pinned contract's `contracts/scanner.yaml`.
//   4. Fails if nothing is discovered (vacuity guard).
//
// Usage
//   node scripts/check-scanner-calls.mjs [--contract <contract-repo-dir>] [--repo <backend-dir>]
//
// Defaults: --contract ../cardtalogue-api-contract, --repo .
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, resolve, sep } from 'node:path'

// ---------------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------------
function parseArgs(argv) {
  const args = { contract: '../cardtalogue-api-contract', repo: '.' }
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a === '--contract') args.contract = argv[++i]
    else if (a === '--repo') args.repo = argv[++i]
    else if (a === '--help' || a === '-h') {
      console.log('usage: node scripts/check-scanner-calls.mjs [--contract <dir>] [--repo <dir>]')
      process.exit(0)
    } else {
      console.error(`unknown argument: ${a}`)
      process.exit(2)
    }
  }
  return args
}

const args = parseArgs(process.argv.slice(2))
const contractDir = resolve(args.contract)
const repoDir = resolve(args.repo)
const SCAN_ENV = 'CARD_SCAN_API_URL'

const log = (...a) => console.log(...a)
const failures = []
const warnings = []

// ---------------------------------------------------------------------------
// Path normalisation
// ---------------------------------------------------------------------------
function normalizePath(p) {
  let out = String(p).trim()
  const q = out.indexOf('?')
  if (q !== -1) out = out.slice(0, q)
  out = out.replace(/\$\{[^}]*\}/g, '{param}')
  if (!out.startsWith('/')) out = '/' + out
  out = out.replace(/\/+$/, '')
  return out === '' ? '/' : out
}

function pathShape(p) {
  return normalizePath(p).replace(/\{[^}]*\}/g, '{}')
}

// ---------------------------------------------------------------------------
// Pinned contract (scanner.yaml) — indentation parse, no YAML dependency
// ---------------------------------------------------------------------------
function loadContractPaths(file) {
  const lines = readFileSync(file, 'utf8').split('\n')
  const paths = new Map()
  let inPaths = false
  let current = null
  for (const raw of lines) {
    const line = raw.replace(/\t/g, '  ')
    if (/^paths:\s*$/.test(line)) {
      inPaths = true
      continue
    }
    if (!inPaths) continue
    if (/^\S/.test(line) && !/^paths:/.test(line)) break
    const pathKey = line.match(/^ {2}(\/[^:]*):\s*$/)
    if (pathKey) {
      current = pathKey[1]
      paths.set(current, new Set())
      continue
    }
    const method = line.match(/^ {4}(get|put|post|delete|options|head|patch|trace):\s*$/)
    if (method && current) paths.get(current).add(method[1].toUpperCase())
  }
  return paths
}

const contractFile = join(contractDir, 'contracts', 'scanner.yaml')
if (!existsSync(contractFile)) {
  console.error(`FAIL: scanner contract not found: ${contractFile}`)
  process.exit(1)
}
const contractPaths = loadContractPaths(contractFile)
if (contractPaths.size === 0) {
  console.error(`FAIL: parsed 0 paths from ${contractFile}`)
  process.exit(1)
}

// ---------------------------------------------------------------------------
// Scanner URL discovery
// ---------------------------------------------------------------------------
const urls = [] // { url, source, priority }
function addUrl(url, source, priority) {
  if (!url) return
  let parsed
  try {
    parsed = new URL(url)
  } catch {
    warnings.push(`ignoring unparsable scanner URL from ${source}: ${url}`)
    return
  }
  if (urls.some((u) => u.url === url)) return
  urls.push({ url, source, priority, host: parsed.host, path: normalizePath(parsed.pathname) })
}

if (process.env[SCAN_ENV]) addUrl(process.env[SCAN_ENV], `environment variable ${SCAN_ENV}`, 0)

const configFile = join(repoDir, 'internal', 'config', 'config.go')
let configSrc = ''
if (existsSync(configFile)) {
  configSrc = readFileSync(configFile, 'utf8')
  const re = new RegExp(`getEnv\\(\\s*"${SCAN_ENV}"\\s*,\\s*"([^"]+)"`)
  const m = configSrc.match(re)
  if (m) addUrl(m[1], `${relative(process.cwd(), configFile)} default`, 1)
}

const makefile = join(repoDir, 'Makefile')
if (existsSync(makefile)) {
  const re = new RegExp(`^[^#\\n]*${SCAN_ENV}\\s*[?:]?=\\s*(\\S+)`, 'm')
  const m = readFileSync(makefile, 'utf8').match(re)
  if (m) addUrl(m[1], `${relative(process.cwd(), makefile)}`, 2)
}

const envFile = join(repoDir, '.env')
if (existsSync(envFile)) {
  const m = readFileSync(envFile, 'utf8').match(new RegExp(`^${SCAN_ENV}=(.+)$`, 'm'))
  if (m) addUrl(m[1].trim(), `${relative(process.cwd(), envFile)}` + ' (not in CI)', 3)
}

if (urls.length === 0) {
  console.error(`FAIL: could not determine the scanner URL (${SCAN_ENV}) from env, .env, config.go or Makefile.`)
  process.exit(1)
}
urls.sort((a, b) => a.priority - b.priority)
const primary = urls[0]
const scannerHosts = new Set(urls.map((u) => u.host))
log(`Scanner URL: ${primary.url}`)
log(`  source: ${primary.source}`)
if (urls.length > 1) {
  for (const u of urls.slice(1)) log(`  also seen: ${u.url}  (${u.source})`)
}

// ---------------------------------------------------------------------------
// Go sources
// ---------------------------------------------------------------------------
function walk(dir, out = []) {
  for (const entry of readdirSync(dir)) {
    if (entry === '.git' || entry === 'vendor' || entry === 'node_modules') continue
    const full = join(dir, entry)
    const st = statSync(full)
    if (st.isDirectory()) walk(full, out)
    else if (entry.endsWith('.go') && !entry.endsWith('_test.go')) out.push(full)
  }
  return out
}

const goFiles = walk(repoDir).sort()
if (goFiles.length === 0) {
  console.error(`FAIL: no Go sources found under ${repoDir}`)
  process.exit(1)
}
const sources = goFiles.map((file) => ({
  file,
  rel: relative(process.cwd(), file).split(sep).join('/'),
  src: readFileSync(file, 'utf8'),
}))
const allSrc = sources.map((s) => s.src).join('\n')

function lineOf(src, index) {
  return src.slice(0, index).split('\n').length
}

// ---------------------------------------------------------------------------
// Identifiers bound to the scanner URL (config field, ctor param, struct field)
// ---------------------------------------------------------------------------
const boundNames = new Set()
const configFieldMatch = configSrc.match(new RegExp(`(\\w+):\\s*getEnv\\(\\s*"${SCAN_ENV}"`))
if (configFieldMatch) boundNames.add(configFieldMatch[1])

function remember(name) {
  if (name) boundNames.add(name)
}

// Identifiers assigned from the config field, then the constructors they feed.
const fromConfig = new Set()
for (const field of boundNames) {
  const re = new RegExp(`(\\w+)\\s*:?=\\s*(?:config|cfg)\\.${field}\\b`, 'g')
  let m
  while ((m = re.exec(allSrc)) !== null) {
    fromConfig.add(m[1])
  }
}
for (const name of fromConfig) remember(name)

// NewX(<ident>) where <ident> is bound -> trace into NewX's params and the
// struct fields it assigns, so `s.apiURL` inside the service is recognised.
for (const name of [...boundNames]) {
  const callRe = new RegExp(`New(\\w+)\\(([^)]*\\b${name}\\b[^)]*)\\)`, 'g')
  let m
  while ((m = callRe.exec(allSrc)) !== null) {
    const ctor = m[1]
    const ctorBodyRe = new RegExp(`func\\s+New${ctor}\\(([^)]*)\\)[^{]*\\{([\\s\\S]*?)\\n\\}`)
    const cm = allSrc.match(ctorBodyRe)
    if (!cm) continue
    const params = cm[1]
      .split(',')
      .map((p) => p.trim().split(/\s+/)[0])
      .filter(Boolean)
    for (const param of params) remember(param)
    const body = cm[2]
    const fieldRe = /(\w+)\s*:\s*(\w+)\s*,/g
    let fm
    while ((fm = fieldRe.exec(body)) !== null) {
      if (params.includes(fm[2])) remember(fm[1])
    }
  }
}

// ---------------------------------------------------------------------------
// Outbound calls → (method, path)
// ---------------------------------------------------------------------------
const calls = [] // { method|null, path, source }
const unmapped = [] // calls to a URL we could not tie to the scanner

function urlForExpr(expr) {
  if (/^["'`]https?:\/\//.test(expr)) return expr.replace(/^["'`]|["'`]$/g, '')
  const tail = expr.split('.').pop()
  if (boundNames.has(tail)) return primary.url
  if (boundNames.has(expr)) return primary.url
  return null
}

const NEW_REQUEST = /http\.NewRequest(WithContext)?\s*\(/g
for (const { rel, src } of sources) {
  NEW_REQUEST.lastIndex = 0
  let m
  while ((m = NEW_REQUEST.exec(src)) !== null) {
    const window = src.slice(m.index, m.index + 400)
    const withCtx = Boolean(m[1])
    // <ctx>, "METHOD", <expr> | <ctx>, http.MethodX, <expr> | "METHOD", <expr>
    const argRe = withCtx
      ? /NewRequestWithContext\(\s*[^,]+,\s*(?:"([A-Z]+)"|http\.Method([A-Za-z]+))\s*,\s*(.+?)(?:\s*\+\s*"([^"]*)")?\s*[,)]/
      : /NewRequest\(\s*(?:"([A-Z]+)"|http\.Method([A-Za-z]+))\s*,\s*(.+?)(?:\s*\+\s*"([^"]*)")?\s*[,)]/
    const am = window.match(argRe)
    if (!am) {
      warnings.push(`could not parse http.NewRequest call at ${rel}:${lineOf(src, m.index)}`)
      continue
    }
    const method = (am[1] || am[2] || '').toUpperCase()
    const expr = am[3].trim()
    const suffix = am[4] || ''
    const where = `${rel}:${lineOf(src, m.index)}`
    const base = urlForExpr(expr)

    if (!base) {
      unmapped.push({ method, expr, where })
      continue
    }
    let basePath
    try {
      basePath = base === primary.url ? primary.path : normalizePath(new URL(base).pathname)
    } catch {
      basePath = normalizePath(base)
    }
    const path = normalizePath(suffix ? basePath + suffix : basePath)
    calls.push({ method, path, where })
  }
}

// Scanner URL literals anywhere in non-test Go code (e.g. config defaults).
for (const { rel, src } of sources) {
  const litRe = /"(https?:\/\/[^"\s]+)"/g
  let m
  while ((m = litRe.exec(src)) !== null) {
    let parsed
    try {
      parsed = new URL(m[1])
    } catch {
      continue
    }
    const isScanner = scannerHosts.has(parsed.host) || parsed.port === '5000'
    if (!isScanner) continue
    calls.push({ method: null, path: normalizePath(parsed.pathname), where: `${rel}:${lineOf(src, m.index)}` })
  }
}

// ---------------------------------------------------------------------------
// Report
// ---------------------------------------------------------------------------
const shapes = new Map()
for (const [path, methods] of contractPaths) {
  const key = pathShape(path)
  if (!shapes.has(key)) shapes.set(key, new Map())
  for (const method of methods) shapes.get(key).set(method, path)
}

log(`Contract: ${contractFile}`)
log(`  ${contractPaths.size} scanner paths\n`)

const seen = new Set()
log('Backend -> scanner calls:')
if (calls.length === 0) log('  (none found)')
// Method-ful calls first so a path-only sighting of the same endpoint collapses
// into it.
calls.sort((a, b) => (a.method ? 0 : 1) - (b.method ? 0 : 1))
const reportedShapes = new Set()
for (const call of calls) {
  const key = `${call.method || '*'} ${call.path}`
  if (seen.has(key)) continue
  const shape = pathShape(call.path)
  if (!call.method && reportedShapes.has(shape)) continue
  seen.add(key)
  reportedShapes.add(shape)
  const entry = shapes.get(shape)
  if (!entry) {
    failures.push(`${key}  (${call.where}) — path not in scanner.yaml`)
    log(`  x ${key}  (${call.where}) — path not in scanner.yaml`)
    continue
  }
  if (call.method && !entry.has(call.method)) {
    const allowed = [...entry.keys()].sort().join('|') || 'none'
    failures.push(`${key}  (${call.where}) — method not in scanner.yaml (allowed: ${allowed})`)
    log(`  x ${key}  (${call.where}) — method not in scanner.yaml (allowed: ${allowed})`)
    continue
  }
  log(`  ok ${key}${call.method ? '' : ' (path only)'}  (${call.where})`)
}

if (unmapped.length > 0) {
  log('\nOutbound http.NewRequest calls not tied to the scanner (informational):')
  for (const u of unmapped) log(`  - ${u.method} ${u.expr}  (${u.where})`)
}
if (warnings.length > 0) {
  log('\nWarnings:')
  for (const w of warnings) log('  ! ' + w)
}

log('\n' + '='.repeat(60))
if (calls.length === 0) {
  console.error('FAIL: discovered 0 scanner calls — the scanner is blind (vacuity guard).')
  process.exit(1)
}
if (failures.length > 0) {
  console.error(`FAIL: ${failures.length} scanner call(s) not covered by the pinned contract.`)
  for (const f of failures) console.error('  - ' + f)
  console.error('Update contracts/scanner.yaml in the contract repo before merging.')
  process.exit(1)
}
console.log(`PASS: all ${seen.size} scanner call(s) are covered by the pinned contract.`)
process.exit(0)
