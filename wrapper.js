const core = require('@actions/core');
const os = require('os');

function execute(binary, system) {
    core.info(`[wrapper] runner platform: '${system}'`);
    core.debug(`[wrapper] arguments: '${process.argv.slice(2).join(" ")}'`);

    require('child_process').execSync(
        `${binary} ${process.argv.slice(2).join(" ")}`,
        { stdio: 'inherit' }
    );

    core.info(`[wrapper] finished`);
}

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