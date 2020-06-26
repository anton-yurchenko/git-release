const core = require('@actions/core');
const os = require('os');

function execute(binary, system) {
    core.info(`[wrapper] runner platform: '${system}'`);
    core.info(`[wrapper] arguments: '${core.getInput('args')}'`);

    require('child_process').execSync(
        `${binary} ${core.getInput('args')}`,
        { stdio: 'inherit' }
    );
}

core.warning(`[wrapper] executing this action via wrapper is not recommended! see README for more information`);

if (os.arch != `x64`) {
    core.setFailed(`[wrapper] runner cpu architecture is not supported: '${os.arch}'`);
}

if (os.type == `Windows_NT`) {
    execute(`${__dirname}\\build\\git-release-windows-amd64.exe`, os.type);
} else if (os.type == `Linux`) {
    execute(`${__dirname}/build/git-release-linux-amd64`, os.type);
} else {
    core.setFailed(`[wrapper] runner operation system is not supported: '${os.type}'`);
}