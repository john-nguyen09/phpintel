const App = require('./lib/app').App;
const path = require('path');
const Hasher = require('./lib/service/hasher').Hasher;
const Indexer = require('./lib/index/indexer').Indexer;
const homedir = require('os').homedir();

const rootPath = 'C:\\Users\\johnn\\Development\\MindAtlas\\LMS\\benevolent';
const hasher = new Hasher();

const storagePath = path.join(homedir, '.phpintel', hasher.getHash(rootPath));
App.init(storagePath, null);
console.time('indexing');

const indexer = App.get(Indexer);
indexer.indexWorkspace(rootPath)
    .then(() => {
        console.timeEnd('indexing');
    });
