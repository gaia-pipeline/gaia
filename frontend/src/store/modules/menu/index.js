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
      name: 'Vault',
      path: '/vault',
      meta: {
        icon: 'fa-lock'
      },
      component: lazyLoading('vault', true)
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
      name: 'Permissions',
      path: '/permissions',
      meta: {
        icon: 'fa-users'
      },
      component: lazyLoading('permissions', true)
    }
  ]
}

export default {
  state
}
