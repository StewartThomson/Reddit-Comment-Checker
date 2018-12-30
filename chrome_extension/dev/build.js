const fs = require('fs-extra');
const concat = require('concat');
const glob = require("glob");
const path = require("path");

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
    const icons = glob.sync("./src/icons/*.png");
    if (!fs.existsSync('../dist/icons')){
        fs.mkdirSync('../dist/icons');
    }
    for(let icon of icons) {
        let filename = path.basename(icon);
        await fs.copyFile(
            icon,
            '../dist/icons/' + filename
        )
    }
})();