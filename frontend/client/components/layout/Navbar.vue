<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar">
        <div class="navbar-start">
          <div class="search-icon">
            <i class="fa fa-search fa-lg" aria-hidden="true"/>
          </div>
          <div>
            <input class="borderless-search" type="text" placeholder="Find pipeline ..." v-model="search">
          </div>
          <div class="navbar-button">
            <a class="button is-primary" @click="createPipeline">
              <span class="icon">
                <i class="fa fa-plus"></i>
              </span>
              <span>Create Pipeline</span>
            </a>
          </div>
        </div>
        <div class="navbar-end">
          <a class="navbar-item signed-text" v-if="session">
            <span>Hi, {{ session.display_name }}</span>
            <div class="avatar">
              <svg class="avatar-img" data-jdenticon-value="session.display_name"></svg>
            </div>
          </a>
          <a class="navbar-item signed-in-icons-div" @click="refresh" v-if="session">
            <i class="fa fa-refresh fa-lg signed-in-icons" aria-hidden="true"/>
          </a>
          <a class="navbar-item signed-in-icons-div" @click="logout" v-if="session">
            <i class="fa fa-sign-out fa-lg signed-in-icons" aria-hidden="true"/>
          </a>
        </div>
      </nav>
    </div>
  </section>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import auth from '../../auth'
import jdenticon from 'jdenticon'

export default {

  data () {
    return {
      search: ''
    }
  },

  components: {
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
    this.reload()
  },

  watch: {
    '$route': 'reload'
  },

  methods: {
    reload () {
      // Update jdenticon to prevent rendering issues
      jdenticon()
    },

    refresh () {
      window.location.reload()
    },

    logout () {
      this.$router.push('/')
      auth.logout(this)
    },

    createPipeline () {
      this.$router.push('/pipelines/create')
    },

    ...mapActions([
      'toggleSidebar'
    ])
  }
}
</script>

<style lang="scss">

.navbar-button {
  padding-top: 17px;  
}

.avatar {
  margin-left: 10px;
  width: 40px;
  height: 40px;
  overflow: hidden;
  border-radius: 50%;
  position: relative;
  border-color: whitesmoke;
  border-style: solid;
}

.avatar-img {
  position: absolute;
  width: 50px;
  height: 50px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.signed-text {
  color: #8c91a0;
  font-weight: bold;
  text-transform: capitalize;
  border-right: solid 1px #8c91a0;
  padding-right: 30px;
}

.navbar-start {
  padding-left: 240px;
}

.search-icon {
  padding-top: 22px;
  color: whitesmoke;
}

.signed-in-icons {
  padding: 10px;
}

.signed-in-icons-div {
  color: whitesmoke;
}

.borderless-search {
  border: none;
  border-color: transparent;
  width: 300px;
  height: 70px;
  background-color: transparent;
  padding-left: 20px;
  color: whitesmoke;
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

.borderless-search::-moz-placeholder {
    color: #8c91a0;
    text-shadow: none;
    opacity: 1;
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

.sign-in-text {
  font-size: 20px;
  font-weight: bold;
  color: whitesmoke;
  padding-top: 6px;
  padding-right: 10px;
}

.sign-in-icon {
  color: #4da2fc;
  padding-right: 15px;
  padding-top: 7px;
}
</style>
