const core = require('@actions/core');
const os = require('os');
const path = require('path');

function execute(binary, system) {
    core.info(`[wrapper] runner platform: '${system}'`);

    require('child_process').execFileSync(
        binary,
        [core.getInput('args')],
        { stdio: 'inherit' }
    );
}

function main() {
    if (os.arch != `x64`) {
        core.setFailed(`[wrapper] runner cpu architecture is not supported: '${os.arch}'`);
        process.exit(1);
    }

    if (process.platform == `win32`) {
        filename = `git-release-windows-amd64.exe`
    } else if (process.platform == `linux`) {
        core.warning(`Executing this action via wrapper is not recommended on Linux runner!`);
        filename = `git-release-linux-amd64`
    } else {
        core.setFailed(`[wrapper] runner operation system is not supported: '${process.platform}'`);
        process.exit(1);
    }

    file = path.join(__dirname, `build/${filename}`)
    execute(file, process.platform);
}

main()