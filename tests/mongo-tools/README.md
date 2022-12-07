## Requirements 
- data restored from mongodb is the same in FerretDB
  - NOTE: the order of documents may be different
  - ... because of that, the dump file differs

- dump file from FerretDB is the same every time on a same set of data
  - that would prove that dumping/restoring process is correct

## To make sure that above requirements are met we want to (step by step):
- restore the data from the prepared dump A
- get the current database state X (by using mongo-driver)
- dump the data to dump B
- drop the database and restore from the dump B
- get the current state Y
- compare current state Y with the state X
  - this step proves that the restoring and dumping process works **But it didn't prove that the data restored from initial dump was correct**
- compare old dump A file with the new one (dump B)
  - this step proves that the initial restore was for sure valid (it didn't omit any data)

## Dump files problem:
Dump file created by FerretDB differs from the mongodb's one. It's because of a different order of documents in collections (NOTE: do we have any plans to fix that?)

- we checked that FerretDB dump will have the same order every time
- so if we create another dump file, specifically for FerretDB and if we make sure that it contains the same data as MongoDB's one
- ... we would be able to test everything correctly

## To make sure that the dump file doesn't contain invalid data we can:
- While using FerretDB, restore mongodb dump A file and after that, dump the data to the dump B
- While using MongoDB, restore mongodb dump A file
- Set expectedState to the database state after restoring mongodb dump A
- Set actualState to the database state after restoring FerretDB dump B
- compare states
- states for sure are not the same because of a different order in dumps
- to prove that there's no other data inconsistency than the order, the test iterates through every document in both of states and checks
if every single document occures at the same index
  - if indexes are the same, then we're sure that the order for specified documents sequence is the same
  - if indexes are not the same, then we're sure that the order is different
  - if one of above statements is correct, that means that the data in FerretDB dump B is the same as in MongoDB and it can differ ONLY
  by an order
  - ... otherwise, one of the dump contains at least one different document than the other one so we shouldn't allow it


## Summary
- With this solution we have 2 sample dumps in repository (one for MongoDB and one for FerretDB)
- We are sure that both of them contain the same data, and they only differ by an order
- Both of them pass tests for their database and fail for the other one
