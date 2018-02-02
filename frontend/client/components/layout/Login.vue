<template>
  <div class="content has-text-centered login-box-outter">
    <div class="login-box-middle">
      <div class="box login-box-inner">
        <h1 class="title header-text">Gaia</h1>
        <div class="block login-box-content">
          <div class="login-box-content">
            <p class="control has-icons-left">
              <input class="input is-large input-bar" v-focus type="text" v-model="username" @keyup.enter="login" placeholder="Username">
              <span class="icon is-small is-left">
                <i class="fa fa-user-circle"></i>
              </span>
            </p>
          </div>
          <div class="login-box-content">
            <p class="control has-icons-left">
              <input class="input is-large input-bar" type="password" @keyup.enter="login" v-model="password" placeholder="Password">
              <span class="icon is-small is-left">
                <i class="fa fa-lock"></i>
              </span>
            </p>
          </div>
          <div class="login-box-content">
            <button class="button is-primary login-button" @click="login">Sign In</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
import Notification from 'vue-bulma-notification-fixed'
import auth from '../../auth'

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

export default {

  data () {
    return {
      username: '',
      password: ''
    }
  },

  methods: {
    login () {
      var credentials = {
        username: this.username,
        password: this.password
      }

      // Authenticate
      auth.login(this, credentials)
        .then((response) => {
          if (!response) {
            openNotification({
              title: 'Invalid credentials!',
              message: 'Wrong username and/or password.',
              type: 'danger'
            })
          }
        })
    }
  }
}
</script>

<style lang="scss" scoped>
.login-box-outter {
  display: table;
  position: absolute;
  height: 100%;
  width: 100%;
}

.login-box-middle {
  display: table-cell;
  vertical-align: middle;
}

.login-box-inner {
  text-align: center;
  background-color: #3f3d49;
  width: 40%;
  margin-left: auto;
  margin-right: auto;
}

.login-box-content {
  margin: auto;
  padding: 10px;
}

.login-button {
  width: 150px;
  height: 50px;
}

.header-text {
  color: #4da2fc;
  padding-bottom: 15px;
  font-size: 4rem;
}
</style>
