<template>
  <div id="app">
    <nprogress-container></nprogress-container>
    <navbar :show="true"></navbar>
    <sidebar :show="sidebar.opened && !sidebar.hidden"></sidebar>
    <app-main></app-main>
  </div>
</template>

<script>
import NprogressContainer from 'vue-nprogress/src/NprogressContainer'
import { Navbar, Sidebar, AppMain } from 'components/layout/'
import { mapGetters, mapActions } from 'vuex'

export default {
  components: {
    Navbar,
    Sidebar,
    AppMain,
    NprogressContainer
  },

  beforeMount () {
    const { body } = document
    const WIDTH = 768
    const RATIO = 3

    const handler = () => {
      if (!document.hidden) {
        let rect = body.getBoundingClientRect()
        let isMobile = rect.width - RATIO < WIDTH
        this.toggleDevice(isMobile ? 'mobile' : 'other')
        this.toggleSidebar({
          opened: !isMobile
        })
      }
    }

    document.addEventListener('visibilitychange', handler)
    window.addEventListener('DOMContentLoaded', handler)
    window.addEventListener('resize', handler)
  },

  computed: mapGetters({
    sidebar: 'sidebar'
  }),

  methods: mapActions([
    'toggleDevice',
    'toggleSidebar'
  ])
}
</script>

<style lang="scss">
@import '~animate.css';
.animated {
  animation-duration: .377s;
}

@import '~bulma';
@import '~wysiwyg.css/wysiwyg.sass';

$fa-font-path: '~font-awesome/fonts/';
@import '~font-awesome/scss/font-awesome';

html {
  background-color: rgb(42, 38, 53);
}

@font-face {
  font-family: 'Lobster';
  src: url('~assets/Lobster-Regular.ttf');
}

.nprogress-container {
  position: fixed !important;
  width: 100%;
  height: 50px;
  z-index: 2048;
  pointer-events: none;

  #nprogress {
    $color: #48e79a;

    .bar {
      background: $color;
    }
    .peg {
      box-shadow: 0 0 10px $color, 0 0 5px $color;
    }

    .spinner-icon {
      border-top-color: $color;
      border-left-color: $color;
    }
  }
}

.is-primary {
  background-color: #4da2fc !important;
  font-weight: bold;
  border-color: transparent;
  color: whitesmoke;
}

.is-primary:hover, .button:active, .button:focus {
  color: whitesmoke;
  border-color: transparent;
}

.content-article {
  color: whitesmoke;
  background-color: #3f3d49;
}

.label {
  color: whitesmoke;
  font-weight: normal;
}

.input-bar {
  background-color: #19191b;
  color: white;
  border-color: #2a2735;
}

.input-bar::-webkit-input-placeholder {
    color: #8c91a0;
    text-shadow: none;
    -webkit-text-fill-color: initial;
}

.input-bar::-moz-placeholder { 
    color: #8c91a0;
    text-shadow: none;
    opacity: 1;
}

.title-text {
  border-bottom: 1px whitesmoke solid;
  padding-bottom: 8px;
  text-align: center;
  color: #ff651d !important;
}

.modal-z-index {
  z-index: 1025;
}

@media screen and (min-width: 768px) {
  .modal-content {
    width: 480px; /* either % (e.g. 60%) or px (400px) */
  }
}

.collapse-item {
  background-color: black;

  .card-header {
    background-color: #3f3d49;
  }
}

.card-header-title {
  color: whitesmoke;
}

</style>
