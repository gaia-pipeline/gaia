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
          <span class="has-text-white" v-if="session">{{ session.display_name }}</span>
        </div>
      </nav>
    </div>

    <!-- Login modal -->
    <modal :visible="loginModal" class="modal-z-index" @close="close">
          <article class="tile is-child box">
            <h1 class="title">Sign In</h1>
            <div class="block">
              <p class="control has-icons-left">
                <input class="input is-large" type="text" v-model="username" placeholder="Username">
                <span class="icon is-small is-left">
                  <i class="fa fa-user-circle"></i>
                </span>
              </p>
              <p class="control has-icons-left">
                <input class="input is-large" type="password" v-model="password" placeholder="Password">
                <span class="icon is-small is-left">
                  <i class="fa fa-lock"></i>
                </span>
              </p>
              <p class="control">
                <button class="button is-primary" @click="login">Sign In</button>
              </p>
            </div>
          </article> 
    </modal>
  </section>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import { Modal } from 'vue-bulma-modal'
import auth from '../../auth'

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
    Modal
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

.navbar-start {
  padding-left: 240px;
}

.search-icon {
  padding-top: 22px;
  color: whitesmoke;
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
