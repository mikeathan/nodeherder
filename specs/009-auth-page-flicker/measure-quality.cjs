// Run from the repository root after the coverage command in quality.md.
const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const ts = require(path.resolve('frontend/node_modules/typescript'));

if (ts.version !== '5.9.2') throw new Error(`Expected locked TypeScript 5.9.2, got ${ts.version}`);
const sourcePath = path.resolve('frontend/src/composables/useLoginSession.ts');
const coverage = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'))[sourcePath];
if (!coverage) throw new Error('Missing composable coverage');

const isCallable = (node) => ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node);
const decisions = new Set([
  ts.SyntaxKind.IfStatement, ts.SyntaxKind.ConditionalExpression,
  ts.SyntaxKind.ForStatement, ts.SyntaxKind.ForInStatement, ts.SyntaxKind.ForOfStatement,
  ts.SyntaxKind.WhileStatement, ts.SyntaxKind.DoStatement, ts.SyntaxKind.CaseClause,
  ts.SyntaxKind.CatchClause,
]);
const logicalOperators = new Set([
  ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken,
]);

function report(text, fileName, instrumented) {
  const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
  const rows = [];
  function visit(node) {
    if (isCallable(node) && node.body) rows.push(measure(node));
    ts.forEachChild(node, visit);
  }
  function measure(callable) {
    const nestedRanges = [];
    let complexity = 1;
    function count(node) {
      if (isCallable(node)) {
        nestedRanges.push([node.getStart(source), node.end]);
        return;
      }
      if (decisions.has(node.kind)) complexity += 1;
      if (ts.isBinaryExpression(node) && logicalOperators.has(node.operatorToken.kind)) complexity += 1;
      ts.forEachChild(node, count);
    }
    count(callable.body);
    const statements = instrumented ? Object.entries(instrumented.statementMap).filter(([, location]) => {
      const start = source.getPositionOfLineAndCharacter(location.start.line - 1, location.start.column);
      return start >= callable.body.getStart(source) && start < callable.body.end &&
        !nestedRanges.some(([from, to]) => start >= from && start < to);
    }) : [];
    const covered = statements.filter(([id]) => instrumented.s[id] > 0).length;
    const fraction = statements.length ? covered / statements.length : 0;
    const line = source.getLineAndCharacterOfPosition(callable.getStart(source)).line + 1;
    const name = callable.name?.getText(source) ||
      (ts.isVariableDeclaration(callable.parent) ? callable.parent.name.getText(source) : `callback@${line}`);
    return {
      name, line, complexity, statements: statements.length, covered,
      coverage: instrumented && statements.length ? fraction : null,
      crap: Number((complexity ** 2 * (1 - fraction) ** 3 + complexity).toFixed(2)),
      basis: instrumented && statements.length ? 'measured statement coverage' : 'zero-coverage upper bound',
    };
  }
  visit(source);
  return rows;
}

const baseline = execFileSync('git', ['show', '6f299a13:frontend/src/components/auth/LoginPage.vue'], { encoding: 'utf8' });
const script = baseline.split('<script setup lang="ts">')[1].split('</script>')[0];
console.log(JSON.stringify({
  typescript: ts.version,
  baseline: report(script, 'LoginPage.before.ts', null),
  after: report(fs.readFileSync(sourcePath, 'utf8'), sourcePath, coverage),
}, null, 2));
