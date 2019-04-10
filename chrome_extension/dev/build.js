const fs = require('fs-extra');
const concat = require('concat');
const glob = require("glob");
const child_process = require("child_process");

(async function build() {
    fs.ensureDirSync('../dist/icons');
    const SRC_FOLDER = "./src/";
    const DIST_FOLDER = "../dist/";
    const ZIP_NAME = "reddit_comment_checker.zip";
    const FILES_TO_COPY = [
        "style.css",
        "manifest.json",
        "icons/*.png"
    ];

    let files = glob.sync(SRC_FOLDER + "*.js");

    console.log("Merging files...");
    await concat(files, DIST_FOLDER + 'reddit-comment-checker.js');

    console.log("Copying files...");
    for(let globToGet of FILES_TO_COPY) {
        files = glob.sync(SRC_FOLDER + globToGet);
        for(let file of files) {
            let filename = file.slice(file.indexOf(SRC_FOLDER) + SRC_FOLDER.length);
            await fs.copyFile(
                file,
                DIST_FOLDER + filename
            )
        }
    }

    console.log("Removing old zip...");
    fs.removeSync(DIST_FOLDER + ZIP_NAME);

    console.log("Zipping...");
    child_process.execSync(`zip -r ${ZIP_NAME} *`, {
        cwd: DIST_FOLDER
    });
})();