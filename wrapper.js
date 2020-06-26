const core = require('@actions/core');
const os = require('os');

function execute(binary, system) {
    core.info(`[wrapper] runner platform: '${system}'`);
    core.info(`[wrapper] arguments: '${core.getInput('args')}'`);

    require('child_process').execFileSync(
        binary,
        [core.getInput('args')],
        { stdio: 'inherit' }
    );
}

// notify that there is a better way of execution :-)
core.warning(`[wrapper] executing this action via wrapper is not recommended! see README for more information`);

// validate cpu architecture
if (os.arch != `x64`) {
    core.setFailed(`[wrapper] runner cpu architecture is not supported: '${os.arch}'`);
}

// execute correct binary basing on operation system
if (os.type == `Windows_NT`) {
    execute(`${__dirname}\\build\\git-release-windows-amd64.exe`, os.type);
} else if (os.type == `Linux`) {
    execute(`${__dirname}/build/git-release-linux-amd64`, os.type);
} else {
    core.setFailed(`[wrapper] runner operation system is not supported: '${os.type}'`);
}