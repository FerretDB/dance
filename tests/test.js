// Test new PostgreSQL backend compatibility by performing some operations.
//
// This file must be run 3 times in the following order:
// 1. on old backend FERRETDB_POSTGRESQL_OLD=true
// 2. on new backend unset FERRETDB_POSTGRESQL_OLD or FERRETDB_POSTGRESQL_OLD=false
// 3. on old backend FERRETDB_POSTGRESQL_OLD=true

const a = db.a;
const b = db.b;
const c = db.c;
const x = db.x;
const y = db.y;

function start() {
// ensures that do not run A more than once.
if (!a.findOne({a: 1})) {
    assert.eq(false, isNewBackend(), 'first run of A must use old backend');
    a.insert({_id: 1, a: 1});
    runA();
  };

  y.insert({runB: true});
}

if (y.findOne({runB: true})) {
    runB();
}

start();

// 2. run B.
function runB() {
    if (a.findOne({runB: true})) {
        assert.eq(true, isNewBackend(), 'B must use new backend');
      
        jsTestLog('running A on new backend');
      
        // assert A on new backend.
        assertA();
      
        jsTestLog('running B on new backend');
      
        c.insert({_id: 1, a: 1});
        c.createIndex({a: 1});
        assert.eq(2, c.getIndexes().length);
      
        b.update({a: 2}, {$set: {a: 3}});
        assert.eq(3, b.findOne({a: 3}).a);
      
        x.insert({verify: true});
        assert.eq(4, db.getCollectionNames().length);
      
        // skips B.
        a.update({runB: true}, {$set: {runB: false}});
    };
    return;
}

// 1. run A.
function runA() {
  jsTestLog('running A on old backend');

  a.createIndex({a: 1});
  assert.eq(2, a.getIndexes().length);

  a.update({a: 1}, {$set: {a: 2}});
  assert.eq(2, a.findOne({a: 2}).a);

  b.insert({_id: 1, a: 1});
  b.createIndex({a: 1});
  assert.eq(2, b.getIndexes().length);

  b.update({a: 1}, {$set: {a: 2}});
  assert.eq(2, b.findOne({a: 2}).a);

  assert.eq(2, db.getCollectionNames().length);

  let res = assert.commandWorked(db.runCommand({count: 'a'}));
  assert.eq(1, res.n);
  res = assert.commandWorked((db.runCommand({count: 'b'})));
  assert.eq(1, res.n);
  res = assert.commandWorked(db.runCommand({aggregate: 'a', pipeline: [{$project: {a: 1}}, {$count: 'n'}], cursor: {}}));
  assert.eq(1, res.cursor.firstBatch[0].n);
  res = assert.commandWorked(db.runCommand({find: 'a', filter: {}}));
  assert.docEq({_id: 1, a: 2}, res.cursor.firstBatch[0]);
  a.insert({delete: true});
  res = assert.commandWorked(db.runCommand({delete: 'a', deletes: [{q: {delete: true}, limit: 1}]}));
  assert.eq(1, res.n);
  res = assert.commandWorked(db.runCommand({findAndModify: 'a', query: {a: 2}, remove: false, update: {a: 1}}));
  assert.docEq({_id: 1, a: 1}, a.findOne({a: 1}));
  res = assert.commandWorked(db.runCommand({findAndModify: 'a', query: {a: 1}, remove: false, update: {a: 2}}));
  assert.docEq({_id: 1, a: 2}, a.findOne({a: 2}));
  a.insert({runB: true});
  return;
}

// 3. assert B on old backend. END.
if (x.findOne({verify: true})) {
  assert.eq(false, isNewBackend(), 'verify must use old backend');
  jsTestLog('running B on old backend');
  assertB();
  jsTestLog('DONE');
}

function isNewBackend() {
  a.find(); // to get the backend’s version
  let getLog = db.runCommand({getLog: 'startupWarnings'}).log[0];
  getLog = JSON.parse(getLog);
  return isNumeric(getLog.msg[0].slice(-2));
}

function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function assertA() {
  assert.eq(2, a.getIndexes().length);
  assert.eq(2, b.getIndexes().length);
  assert.eq(2, b.findOne({a: 2}).a);
  assert.eq(2, db.getCollectionNames().length);
  return;
}

function assertB() {
  assert.eq(2, a.getIndexes().length);
  assert.eq(2, b.getIndexes().length);
  assert.eq(3, b.findOne({a: 3}).a);
  assert.eq(4, db.getCollectionNames().length);
  return;
}
