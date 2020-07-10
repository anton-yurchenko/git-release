const core = require('@actions/core')
const os = require('os')
const path = require('path')
const fs = require('fs')
const child = require('child_process')

function execute (file) {
  if (fs.existsSync(file)) {
    child.execFileSync(
      file,
      [core.getInput('args')],
      { stdio: 'inherit' }
    )
  } else {
    core.setFailed(`[wrapper] file not found: '${file}'`)
    process.exit(1)
  }
}

function main () {
  if (os.arch() !== 'x64') {
    core.setFailed(`[wrapper] runner cpu architecture is not supported: '${os.arch}'`)
    process.exit(1)
  }

  let filename

  if (process.platform === 'win32') {
    filename = 'git-release-windows-amd64.exe'
  } else if (process.platform === 'linux') {
    core.warning('Executing this action via wrapper is not recommended on Linux runner!')
    filename = 'git-release-linux-amd64'
  } else {
    core.setFailed(`[wrapper] runner operation system is not supported: '${process.platform}'`)
    process.exit(1)
  }

  execute(path.join(__dirname, 'build', filename))
}

main()
