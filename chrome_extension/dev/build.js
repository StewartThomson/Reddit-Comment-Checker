const fs = require('fs-extra');
const concat = require('concat');
const glob = require("glob");

(async function build() {
    const files = glob.sync("./src/*.js");

    await fs.ensureDir('../dist');
    await concat(files, '../dist/reddit-comment-checker.js');
    await fs.copyFile(
        './src/manifest.json',
        '../dist/manifest.json'
    );
    await fs.copyFile(
        './src/style.css',
        '../dist/style.css'
    );
})();