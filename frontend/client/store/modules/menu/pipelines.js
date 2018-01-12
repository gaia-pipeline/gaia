import lazyLoading from './lazyLoading'

export default {
  name: 'Pipelines',
  path: '/pipelines/create',
  meta: {
    icon: 'fa-battery-three-quarters',
    expanded: false
  },
  component: lazyLoading('pipelines/create'),

  subroute: [
    {
      name: 'Create Pipelines',
      path: '/pipelines/create',
      component: lazyLoading('pipelines/create'),
      meta: {
        label: 'Create'
      }
    }
  ]
}
