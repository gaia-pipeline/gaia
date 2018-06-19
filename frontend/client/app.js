import Vue from 'vue'
import axios from 'axios'
import NProgress from 'vue-nprogress'
import { sync } from 'vuex-router-sync'
import App from './App.vue'
import router from './router'
import store from './store'
import * as filters from './filters'
import { TOGGLE_SIDEBAR } from 'vuex-store/mutation-types'
import Notification from 'vue-bulma-notification-fixed'
import auth from './auth'
import lodash from 'lodash'
import VueLodash from 'vue-lodash'

Vue.prototype.$http = axios
Vue.axios = axios
Vue.router = router
Vue.use(NProgress)
Vue.use(VueLodash, lodash)

// Auth interceptors
axios.interceptors.request.use(function (request) {
  request.headers['Authorization'] = 'Bearer ' + auth.getToken()
  return request
})

// Enable devtools
Vue.config.devtools = true

sync(store, router)

const nprogress = new NProgress({ parent: '.nprogress-container' })

const { state } = store

Vue.directive('focus', {
  // When the bound element is inserted into the DOM...
  inserted: function (el) {
    // Focus the element
    el.focus()
  }
})

const NotificationComponent = Vue.extend(Notification)
const openNotification = (propsData = {
  title: '',
  message: '',
  type: '',
  direction: '',
  duration: 4500,
  container: '.notifications'
}) => {
  return new NotificationComponent({
    el: document.createElement('div'),
    propsData
  })
}
Vue.prototype.$notify = openNotification

function handleError (error) {
  // if the server gave a response message, print that
  if (error.response.data.error) {
    // duration should be proportional to the error message length
    openNotification({
      title: 'Error: ' + error.response.status,
      message: error.response.data.error,
      type: 'danger',
      duration: error.response.data.error.length > 60 ? 20000 : 4500
    })
    console.log(error.response.data.error)
  } else {
    if (error.response.status === 404) {
      openNotification({
        title: 'Error: 404',
        message: 'Not found',
        type: 'danger'
      })
    } else if (error.response.status === 403) {
    // Access denied
      openNotification({
        title: 'Error: 403',
        message: 'Not authorized. Please login first.',
        type: 'danger'
      })
    } else {
      openNotification({
        title: 'Error: ' + error.response.status.toString(),
        message: error.response.data,
        type: 'danger'
      })
    }
    console.log(error.response.data)
  }
}
Vue.prototype.$onError = handleError

router.beforeEach((route, redirect, next) => {
  if (state.app.device.isMobile && state.app.sidebar.opened) {
    store.commit(TOGGLE_SIDEBAR, false)
  }
  next()
})

Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key])
})

const app = new Vue({
  router,
  store,
  nprogress,
  ...App
})

export { app, router, store }
