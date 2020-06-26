const core = require('@actions/core');
const os = require('os');

function execute(binary, system) {
    core.info(`[wrapper] runner platform: '${system}'`);

    require('child_process').execFileSync(
        binary,
        [core.getInput('args')],
        { stdio: 'inherit' }
    );
}

if (os.arch != `x64`) {
    core.setFailed(`[wrapper] runner cpu architecture is not supported: '${os.arch}'`);
}

if (os.type == `Windows_NT`) {
    execute(`${__dirname}\\build\\git-release-windows-amd64.exe`, os.type);
} else if (os.type == `Linux`) {
    core.warning(`Executing this action via wrapper is not recommended on Linux runner!`);

    execute(`${__dirname}/build/git-release-linux-amd64`, os.type);
} else {
    core.setFailed(`[wrapper] runner operation system is not supported: '${os.type}'`);
}