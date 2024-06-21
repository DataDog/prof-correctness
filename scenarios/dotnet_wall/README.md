# Description

A dotnet test to check that wall time is captured
Queries are stored in the queries file.

We have two types of thread, one that is quick and the other that takes 1 second extra.
The aim is to check that both were able to process 2 queries and that the wall time adds up to the expected.

# Going further

We can make this much richer. The test also has CPU versus wall time behaviours.
We can also test the grammar from dotnet (with class / function)
