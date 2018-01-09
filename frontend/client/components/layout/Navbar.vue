<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar">
        <div class="navbar-start">
          <div class="search-icon">
            <i class="fa fa-search fa-lg" aria-hidden="true"/>
          </div>
          <div>
            <input class="borderless-search" type="text" placeholder="Find a pipeline ..." v-model="search">
          </div>
        </div>
        <div class="navbar-end">
          <a class="navbar-item" v-if="session === null" v-on:click="showLoginModal">
            <i class="fa fa-sign-in fa-2x sign-in-icon" aria-hidden="true"/>
            <span class="sign-in-text">Sign in</span>
          </a>
          <a class="navbar-item signed-text" v-if="session">
            <span>Hi, {{ session.display_name }}</span>
            <div class="avatar">
              <svg width="40" height="40" data-jdenticon-value="session.display_name"></svg>
            </div>
          </a>
          <a class="navbar-item" v-if="session">
            <i class="fa fa-refresh fa-lg signed-in-icons" aria-hidden="true"/>
          </a>
          <a class="navbar-item" v-if="session">
            <i class="fa fa-sign-out fa-lg signed-in-icons" aria-hidden="true"/>
          </a>
        </div>
      </nav>
    </div>

    <!-- Login modal -->
    <modal :visible="loginModal" class="modal-z-index" @close="close">
      <div class="box login-modal">
        <h1 class="title header-text" style="padding-bottom: 20px;">Sign In</h1>
        <div class="block login-modal-content">
          <div class="login-modal-content">
            <p class="control has-icons-left">
              <input class="input is-large login-input" v-focus type="text" v-model="username" @keyup.enter="login" placeholder="Username">
              <span class="icon is-small is-left">
                <i class="fa fa-user-circle"></i>
              </span>
            </p>
          </div>
          <div class="login-modal-content">
            <p class="control has-icons-left">
              <input class="input is-large login-input" type="password" @keyup.enter="login" v-model="password" placeholder="Password">
              <span class="icon is-small is-left">
                <i class="fa fa-lock"></i>
              </span>
            </p>
          </div>
          <div class="login-modal-content">
            <button class="button is-primary login-button" @click="login">Sign In</button>
          </div>
        </div>
      </div> 
    </modal>
  </section>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import { Modal } from 'vue-bulma-modal'
import auth from '../../auth'
import jdenticon from 'jdenticon'

export default {

  data () {
    return {
      loginModal: false,
      username: '',
      password: '',
      search: ''
    }
  },

  components: {
    Modal,
    jdenticon
  },

  props: {
    show: Boolean
  },

  computed: mapGetters({
    session: 'session',
    pkginfo: 'pkg',
    sidebar: 'sidebar'
  }),

  mounted () {
    this.fetchData()
  },

  watch: {
    '$route': 'fetchData'
  },

  methods: {
    fetchData () {
      let session = auth.getSession()
      if (session) {
        this.$store.commit('setSession', session)
      }
    },

    login () {
      var credentials = {
        username: this.username,
        password: this.password
      }

      auth.login(this, credentials)
      this.close()
    },

    showLoginModal () {
      this.loginModal = true
    },

    close () {
      this.loginModal = false
      this.$emit('close')
    },

    ...mapActions([
      'toggleSidebar'
    ])
  }
}
</script>

<style lang="scss">

.avatar {
  margin-left: 10px;
}

.signed-text {
  color: #8c91a0;
  font-weight: bold;
  text-transform: capitalize;
  border-right: solid 1px #8c91a0;
  padding-right: 30px;
}

.login-modal {
  text-align: center;
  background-color: #2a2735;
}

.login-modal-content {
  margin: auto;
  padding: 10px;
}

.login-input {
  background-color: #3f3d49;
  color: white;
  border-color: #2a2735;
}

.login-input::-webkit-input-placeholder {
    color: #8c91a0;
    text-shadow: none;
    -webkit-text-fill-color: initial;
}

.login-button {
  background-color: #4da2fc !important;
  width: 150px;
  height: 50px;
}

@media screen and (min-width: 768px) {
  .modal-content {
    width: 480px; /* either % (e.g. 60%) or px (400px) */
  }
}

.navbar-start {
  padding-left: 240px;
}

.search-icon {
  padding-top: 22px;
  color: whitesmoke;
}

.signed-in-icons {
  color: whitesmoke;
  padding: 10px;
}

.borderless-search {
  border: none;
  border-color: transparent;
  width: 300px;
  height: 70px;
  background-color: transparent;
  padding-left: 20px;
  color: #4da2fc;
  font-size: 20px;
}

.borderless-search:hover, .borderless-search:focus, .borderless-search:active {
  border: 0;
  border-style: none;
  border-color: transparent;
  outline: none;
}

.borderless-search::-webkit-input-placeholder {
    color: #8c91a0;
    text-shadow: none;
    -webkit-text-fill-color: initial;
}

.app-navbar {
  position: static;
  min-width: 100%;
  height: 70px;
  z-index: 1024 - 1;
  box-shadow: 0 8px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
  background: rgb(60, 57, 74);

  .container {
    margin: auto 10px;
  }
}

.modal-z-index {
  z-index: 1025;
}

.sign-in-text {
  font-size: 20px;
  font-weight: bold;
  color: whitesmoke;
  padding-top: 7px;
}

.sign-in-icon {
  color: #4da2fc;
  padding-right: 15px;
  padding-top: 7px;
}
</style>
