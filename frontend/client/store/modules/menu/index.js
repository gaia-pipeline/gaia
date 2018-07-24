import lazyLoading from './lazyLoading'

// show: meta.label -> name
// name: component name
// meta.label: display label

const state = {
  items: [
    {
      name: 'Overview',
      path: '/overview',
      meta: {
        icon: 'fa-th'
      },
      component: lazyLoading('overview', true)
    },
    {
      name: 'Create Pipeline',
      path: '/pipeline/create',
      meta: {
        icon: 'fa-plus'
      },
      component: lazyLoading('pipeline/create')
    },
    {
      name: 'Settings',
      path: '/settings',
      meta: {
        icon: 'fa-cogs'
      },
      component: lazyLoading('settings', true)
    },
    {
      name: 'Vault',
      path: '/vault',
      meta: {
        icon: 'fa-lock'
      },
      component: lazyLoading('vault', true)
    }
  ]
}

export default {
  state
}
