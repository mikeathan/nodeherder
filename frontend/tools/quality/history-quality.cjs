// TypeScript 5.9.2 AST complexity; Vue script blocks retain original line positions.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const ts = require('typescript');
const root = path.resolve(__dirname, '../../..');
const base = '771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07';
if (ts.version !== '5.9.2') throw new Error('Use locked TypeScript 5.9.2');
const coverage = process.argv[2] ? JSON.parse(fs.readFileSync(process.argv[2], 'utf8')) : {};
const changed = cp
  .execFileSync('git', ['diff', '--name-only', base], { cwd: root, encoding: 'utf8' })
  .trim()
  .split('\n');
const added = cp
  .execFileSync('git', ['ls-files', '--others', '--exclude-standard'], { cwd: root, encoding: 'utf8' })
  .trim()
  .split('\n');
const files = [...new Set([...changed, ...added])].filter(
  (f) => /^frontend\/(src|tools\/server)\//.test(f) && /\.(ts|vue|mjs)$/.test(f) && !/(__tests__|\.test\.)/.test(f)
);
const calls = (n) =>
  ts.isFunctionDeclaration(n) || ts.isFunctionExpression(n) || ts.isArrowFunction(n) || ts.isMethodDeclaration(n);
const decisions = new Set([
  ts.SyntaxKind.IfStatement,
  ts.SyntaxKind.ConditionalExpression,
  ts.SyntaxKind.ForStatement,
  ts.SyntaxKind.ForInStatement,
  ts.SyntaxKind.ForOfStatement,
  ts.SyntaxKind.WhileStatement,
  ts.SyntaxKind.DoStatement,
  ts.SyntaxKind.CaseClause,
  ts.SyntaxKind.CatchClause,
]);
const operators = new Set([
  ts.SyntaxKind.AmpersandAmpersandToken,
  ts.SyntaxKind.BarBarToken,
  ts.SyntaxKind.QuestionQuestionToken,
]);
function script(text, filename) {
  if (!filename.endsWith('.vue')) return text;
  const match = /<script[^>]*>([\s\S]*?)<\/script>/.exec(text);
  if (!match) return '';
  return '\n'.repeat(text.slice(0, match.index).split('\n').length - 1) + match[1];
}
function measure(text, filename, instrumented) {
  const source = ts.createSourceFile(filename, script(text, filename), ts.ScriptTarget.Latest, true);
  const rows = [];
  function visit(node) {
    if (calls(node) && node.body) rows.push(score(node));
    ts.forEachChild(node, visit);
  }
  function score(fn) {
    let complexity = 1;
    const nested = [];
    function count(n) {
      if (calls(n)) {
        nested.push([n.getStart(source), n.end]);
        return;
      }
      if (decisions.has(n.kind)) complexity++;
      if (ts.isBinaryExpression(n) && operators.has(n.operatorToken.kind)) complexity++;
      ts.forEachChild(n, count);
    }
    count(fn.body);
    const statements = Object.entries(instrumented?.statementMap ?? {}).filter(([, loc]) => {
      const pos = source.getPositionOfLineAndCharacter(loc.start.line - 1, loc.start.column);
      return pos >= fn.body.getStart(source) && pos < fn.body.end && !nested.some(([a, b]) => pos >= a && pos < b);
    });
    const covered = statements.filter(([id]) => instrumented.s[id] > 0).length;
    const fraction = statements.length ? covered / statements.length : 0;
    const line = source.getLineAndCharacterOfPosition(fn.getStart(source)).line + 1;
    const name =
      fn.name?.getText(source) ??
      (ts.isVariableDeclaration(fn.parent) ? fn.parent.name.getText(source) : `callback@${line}`);
    return {
      name,
      line,
      complexity,
      coverage: statements.length ? fraction : null,
      crap: Number((complexity ** 2 * (1 - fraction) ** 3 + complexity).toFixed(2)),
      basis: statements.length ? 'Istanbul statement coverage' : 'zero-coverage upper bound',
      text: fn.getText(source),
    };
  }
  visit(source);
  return rows;
}
const reports = files.map((file) => {
  let before = '';
  try {
    before = cp.execFileSync('git', ['show', `${base}:${file}`], {
      cwd: root,
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'ignore'],
    });
  } catch {}
  const baseline = measure(before, file);
  const current = fs.existsSync(path.join(root, file))
    ? measure(fs.readFileSync(path.join(root, file), 'utf8'), file, coverage[path.join(root, file)])
    : [];
  const old = new Set(baseline.map((r) => r.text));
  const touched = current.filter((r) => !old.has(r.text));
  const clean = (r) => {
    const row = { ...r };
    delete row.text;
    return row;
  };
  return { file, baseline: baseline.map(clean), touched: touched.map(clean) };
});
const failures = reports.flatMap((r) =>
  r.touched.filter((f) => f.complexity > 10 || f.crap > 30).map((f) => ({ file: r.file, ...f }))
);
process.stdout.write(JSON.stringify({ base, typescript: ts.version, reports, failures }, null, 2) + '\n');
if (failures.length) process.exitCode = 1;
