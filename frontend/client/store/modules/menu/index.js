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
        icon: 'fa-th',
        link: 'overview.vue'
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
    }
  ]
}

export default {
  state
}
