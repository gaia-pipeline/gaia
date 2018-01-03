<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar">
        <div class="nav-center">
          <a class="navbar-item" href="/">
            <img src="~assets/logo.png" :alt="pkginfo.description">
            &nbsp;&nbsp;<div class="header-text">Gaia</div>
          </a>
        </div>
      </nav>
    </div>
  </section>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {

  props: {
    show: Boolean
  },

  computed: mapGetters({
    pkginfo: 'pkg',
    sidebar: 'sidebar'
  }),

  methods: {
    ...mapActions([
      'toggleSidebar'
    ])
  }
}
</script>

<style lang="scss">

@font-face {
  font-family: 'Lobster';
  src: url('~assets/Lobster-Regular.ttf');
}

.header-text {
  font-family: 'Lobster';
  font-size: 2rem;
  color: #4da2fc;
}

.app-navbar {
  position: fixed;
  min-width: 100%;
  z-index: 1024;
  box-shadow: 0 8px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
  background: rgb(60, 57, 74);

  .container {
    margin: auto 10px;
  }

  .nav-right {
    align-items: stretch;
    align-items: stretch;
    flex: 1;
    justify-content: flex-end;
    overflow: hidden;
    overflow-x: auto;
    white-space: nowrap;
  }

  .nav-center {
    align-items: stretch;
    justify-content: center;
    flex-grow: 0;
    flex-shrink: 0;
    margin-left: auto;
    margin-right: auto;
  }
}

.hero-brand {
  .vue {
    margin-left: 10px;
    color: #36AC70;
  }
  .admin {
    color: #28374B;
  }
}
</style>
