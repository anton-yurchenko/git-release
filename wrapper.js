import * as core from '@actions/core'
import os from 'node:os'
import path from 'node:path'
import fs from 'node:fs'
import child from 'node:child_process'

// The binary reads every setting from the environment, including the asset
// list, which the runner has already exported as INPUT_ASSETS. Nothing is
// passed on argv: v6 forwarded the asset list as a single argument here while
// the container received it pre-split by the runner, so the two distribution
// modes disagreed about which assets to upload.
function execute (file) {
  if (fs.existsSync(file)) {
    child.execFileSync(
      file,
      [],
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

  execute(path.join(import.meta.dirname, 'bin', filename))
}

main()
