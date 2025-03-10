async function buildTime() {
    const r = await fetch("/__livereload", { cache: "no-cache"})
    return r.text()
  }

  async function liveReload() {
    console.log("running in development mode")

    let t_old = await buildTime()
    setInterval(async () => {
      let t_new = await buildTime()
      if (t_old !== t_new) {
        // wait until server is ready to reply
        setTimeout(() => window.location.reload(), 1_000)
      }
    }, 1000)
  }

  liveReload().catch(console.error)