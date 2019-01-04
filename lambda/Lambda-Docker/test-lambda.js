const dockerLambda = require('docker-lambda');
const glob = require('glob');
const fs = require('fs-extra');

(function test() {
    const TEST_FOLDER = "./tests/";
    const FUNCTION_FOLDER = "../functions/comment_analyzer";
    fs.ensureDirSync(TEST_FOLDER);

    let event = JSON.parse(fs.readFileSync("api-gateway-framework.json"));

    let files = glob.sync(TEST_FOLDER + "*.json");

    for(let file of files) {
        console.log(`-- Testing ${file} --`);

        event.body = JSON.stringify(require(file));

        console.log(event);

        let result = dockerLambda({
            event: event,
            taskDir: FUNCTION_FOLDER,
            dockerImage: 'lambci/lambda:go1.x',
            handler: 'main',
            dockerArgs: ["-e", "CORS_ORIGIN='*'"]
        });

        console.log(`-- Result: ${result} --`)
    }
})();