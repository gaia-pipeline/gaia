const pkg = state => state.pkg
const app = state => state.app
const device = state => state.app.device
const sidebar = state => state.app.sidebar
const effect = state => state.app.effect
const menuitems = state => state.menu.items
const session = state => state.session
const intervals = state => state.intervals

export {
  pkg,
  app,
  device,
  sidebar,
  effect,
  menuitems,
  session,
  intervals
}
