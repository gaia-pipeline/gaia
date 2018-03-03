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
      path: '/pipelines/create',
      meta: {
        icon: 'fa-plus'
      },
      component: lazyLoading('pipelines/create')
    }
  ]
}

export default {
  state
}
