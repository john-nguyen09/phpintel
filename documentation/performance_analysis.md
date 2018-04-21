# Performance Analysis
This page lists all findings and analysis to help achieve good performance for the language server
## Indexing
### Problem
Current existing [PHP Language Server](https://github.com/felixfbecker/php-language-server) has a major issue in performance, in my analysis one of the problems is caused by indexing.

### Analysis
In PHP, there is no real parralel processing eventhough IO concurrency can be achieved but it is still single-thread process, so when the file is being indexed other processes are stopped. Therefore, in order to have a good performance the indexing process should be simple and straightforward.

### Solution
A solution that I can think of for now is that only traverse the AST once to extract useful information, as this process is expensive. After that any type resolving to Fully qualified name (FQN) is processed so that there will be no going back up the AST tree.

To further improve this, real parrelism should be utilised. PHP provides an external plugin for threading called pthreads but this requires extra installation therefore it should be optional.