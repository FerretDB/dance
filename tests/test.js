// Tests new PostgreSQL backend compatibility by performing some CRUD operations
// and validating the data.

const a = db.a;
const b = db.b;
const c = db.c;
const x = db.x;

// insert once
if (!a.findOne({a: 1})) {
    a.insert({a: 1});
}

if (a.findOne({useNewBackend: true})) {
    // assert A.
    assertA();

    // run B.
    c.insert({a: 1});
    c.createIndex({a: 1});
    assert.eq(2, c.getIndexes().length);

    b.update({a: 2}, {$set: {a: 3}});
    assert.eq(3, b.findOne({a: 3}).a);

    x.insert({runAssert: true});
    assert.eq(4, db.getCollectionNames().length);

    a.update({useNewBackend: true}, {$set: {useNewBackend: false}});
}

// run A.
a.createIndex({a: 1});
assert.eq(2, a.getIndexes().length);

a.update({a: 1}, {$set: {a: 2}});
assert.eq(2, a.findOne({a: 2}).a);

b.insert({a: 1});
b.createIndex({a: 1});
assert.eq(2, b.getIndexes().length);

a.update({a: 1}, {$set: {a: 2}});
assert.eq(2, b.findOne({a: 2}).a);

assert.eq(2, db.getCollectionNames().length);

a.insert({useNewBackend: true});

if (x.findOne({useNewBackend: false})) {
    // assert B on old backend to verify compatibility.
    assertB();
}

function assertA() {
    assert.eq(2, a.getIndexes().length);
    assert.eq(2, b.getIndexes().length);
    assert.eq(2, b.findOne({a: 2}).a);
    assert.eq(2, db.getCollectionNames().length);
}

function assertB() {
    assert.eq(2, a.getIndexes().length);
    assert.eq(2, b.getIndexes().length);
    assert.eq(3, b.findOne({a: 3}).a);
    assert.eq(4, db.getCollectionNames().length);
}
