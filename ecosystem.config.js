module.exports = {
  apps : [{
    name: "golang-rekap",
    script: "go",
    args: "run .",
    watch: false,
    autorestart: true
  }]
}
