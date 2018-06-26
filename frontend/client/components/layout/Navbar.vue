<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar">
        <div class="navbar-start">
          <div class="search-icon">
            <i class="fa fa-search fa-lg" aria-hidden="true"/>
          </div>
          <div>
            <input class="borderless-search" type="text" placeholder="Find pipeline ..." v-model="search" @input="onChange"
                   @keydown.down="onArrowDown" @keydown.up="onArrowUp" @keydown.enter="onEnter" v-on:keydown.up.prevent>
            <ul id="autocomplete-results" v-show="resultsOpen" class="autocomplete-results" :style="{ height: (results.length * 88.5) + 'px' }">
              <li class="autocomplete-result" v-for="(result, i) in results" :key="i" @click="routeDetailView(result.p)"
                  :class="{ 'is-active': i === arrowCounter }">
                <div class="results-bg">
                <div class="box box-bg">
                  <article class="media">
                    <div class="media-left">
                      <div class="avatar">
                        <img :src="getImagePath(result.p.type)">
                      </div>
                    </div>
                    <div class="media-content" style="margin-top: 7px;">
                      {{ result.p.name }}
                    </div>
                  </article>
                </div>
                </div>
              </li>
            </ul>
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
      search: '',
      resultsOpen: false,
      results: [],
      arrowCounter: 0
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
    document.addEventListener('click', this.handleClickOutside)
  },

  destroyed () {
    document.removeEventListener('click', this.handleClickOutside)
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

    onChange () {
      // reset our arrow counter
      this.arrowCounter = 0

      // Skip empty searches
      if (!this.search) {
        this.resultsOpen = false
        return
      }

      // Get pipelines and search
      this.$http
        .get('/api/v1/pipeline/latest', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            var pipelines = response.data

            // Search
            this.results = pipelines.filter(item => {
              return item.p.name.toLowerCase().indexOf(this.search.toLowerCase()) > -1
            })

            // Open results view
            if (this.results.length > 0) {
              this.resultsOpen = true
            } else {
              this.resultsOpen = false
            }
          }
        })
        .catch((error) => {
          this.$onError(error)
        })
    },

    handleClickOutside (evt) {
      if (!this.$el.contains(evt.target)) {
        this.resultsOpen = false
        this.arrowCounter = 0
      }
    },

    getImagePath (type) {
      return require('assets/' + type + '.png')
    },

    logout () {
      this.$router.push('/')
      this.$store.commit('clearIntervals')
      auth.logout(this)
    },

    createPipeline () {
      this.$router.push('/pipeline/create')
    },

    routeDetailView (pipeline) {
      // Empty search and close results view
      this.resultsOpen = false
      this.search = ''
      this.arrowCounter = 0

      this.$router.push({path: '/pipeline/detail', query: { pipelineid: pipeline.id }})
    },

    onArrowDown () {
      if (this.arrowCounter < (this.results.length - 1)) {
        this.arrowCounter = this.arrowCounter + 1
      }
    },

    onArrowUp () {
      if (this.arrowCounter > 0) {
        this.arrowCounter = this.arrowCounter - 1
      }
    },

    onEnter () {
      // Redirect to view
      this.routeDetailView(this.results[this.arrowCounter].p)

      this.arrowCounter = -1
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
  position: fixed;
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

.autocomplete-results {
  position: absolute;
  z-index: 2500;
  padding: 0;
  margin: 0;
  left: 210px;
  max-width: 350px;
  box-shadow: 20px 0 30px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(255, 255, 255, 0.19);
  height: 177px;
  overflow: auto;
  width: 350px;
  background-color: #2a2735 !important;

}

.autocomplete-result {
  list-style: none;
  text-align: left;
  padding: 4px 6px;
  cursor: pointer;
}

.autocomplete-result.is-active,
.autocomplete-result:hover {
  background-color: #51a0f6;
  color: black;
}

.box-bg {
  background-color: #3f3d49;
  color: whitesmoke;
  text-align: center;
}
</style>
