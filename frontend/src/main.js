import Vue from 'vue'
import axios from 'axios'
import NProgress from 'vue-nprogress'
import { sync } from 'vuex-router-sync'
import App from './App.vue'
import router from './router'
import store from './store'
import * as filters from './filters'
import Notification from 'vue-bulma-notification-fixed'
import auth from './auth'
import lodash from 'lodash'
import VueLodash from 'vue-lodash'

const axiosInstance = axios.create()

Vue.config.productionTip = false
Vue.prototype.$http = axiosInstance
Vue.axios = axiosInstance
Vue.router = router
Vue.use(NProgress, {
  http: false,
  router: false
})
Vue.use(VueLodash, lodash)

// Auth interceptors
axiosInstance.interceptors.request.use(function (request) {
  request.headers['Authorization'] = 'Bearer ' + auth.getToken()
  return request
})

// Enable devtools
Vue.config.devtools = true
sync(store, router)

const nprogress = new NProgress({ parent: '.nprogress-container' })
axiosInstance.interceptors.request.use(function (config) {
  if (!config.params || config.params.hideProgressBar !== true) {
    nprogress.start()
  }
  return config
})
axiosInstance.interceptors.response.use(function (response) {
  nprogress.done()
  return response
})

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
  if (error.response) {
    // duration should be proportional to the error message length
    openNotification({
      title: 'Error: ' + error.response.status,
      message: error.response.data,
      type: 'danger'
    })
  } else if (error.request) {
    openNotification({
      title: 'Error: No response received!',
      message: error.request,
      type: 'danger'
    })
  } else {
    openNotification({
      title: 'Error: Cannot setup request!',
      message: error.message,
      type: 'danger'
    })
  }

  // Finish progress bar
  nprogress.done()
}

Vue.prototype.$onError = handleError

Vue.prototype.$onSuccess = (title, message) => {
  openNotification({
    title: title,
    message: message,
    type: 'success',
    duration: message > 60 ? 20000 : 4500
  })
}

Vue.prototype.$prettifyTags = (tags) => {
  let prettyTags = ''
  for (let i = 0; i < tags.length; i++) {
    if (i === (tags.length - 1)) {
      prettyTags += tags[i]
    } else {
      prettyTags += tags[i] + ', '
    }
  }
  return prettyTags
}

Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key])
})

const app = new Vue({
  router,
  store,
  nprogress,
  ...App
}).$mount('#app')

// A simple event bus
export const EventBus = new Vue()

export { app, router, store }
