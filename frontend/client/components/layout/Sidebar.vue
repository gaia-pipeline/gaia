<template>
  <aside class="menu app-sidebar animated" :class="{ slideInLeft: show, slideOutLeft: !show }">
    <div>
      <a class="navbar-item brand-top" href="/">
        <img src="~assets/logo.png">
        &nbsp;&nbsp;<div class="header-text">Gaia</div>
      </a>
    </div>
    <ul class="menu-list">
      <li v-for="(item, index) in menu" :key="index">
        <router-link :to="item.path" :exact="true" v-if="item.path">
          <span class="icon icon-left is-small"><i :class="['fa', item.meta.icon]"></i></span>
          {{ item.meta.label || item.name }}
        </router-link>
      </li>
    </ul>
  </aside>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  props: {
    show: Boolean
  },

  computed: mapGetters({
    menu: 'menuitems'
  })
}
</script>

<style lang="scss">
@import '~bulma/sass/utilities/mixins';

.header-text {
  font-family: 'Lobster', 'Times', 'serif';
  font-size: 2rem;
  color: #4da2fc;
}

a.navbar-item:hover {
  background-color: #2a2735;
}

.app-sidebar {
  position: fixed;
  top: 0px;
  left: 0;
  padding: 0px 0px 50px;
  width: 200px;
  min-width: 175px;
  max-height: 100vh;
  height: 100%;
  z-index: 1024;
  background: rgb(60, 57, 74);
  box-shadow: 20px 0 30px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
  overflow-y: auto;
  overflow-x: hidden;

  @include mobile() {
    transform: translate3d(-240px, 0, 0);
  }

  .icon {
    vertical-align: baseline;
  }

  .brand-top {
    margin: auto;
    width: 200px;
    padding-left: 45px;
    padding-bottom: 40px;
  }

  .icon-left {
    position: absolute;
    left: 20px;
    margin-top: 13px;
  }

  .menu-list-expanded {
    li a {
      width: 124px !important;
    }
  }

  .menu-list {
    margin: auto;
    width: 200px;

    li {
      float: left;
    }

    li a.is-active {
      background-color: rgb(60, 57, 74);
      color: #51a0f6;
      border-right: 3px solid #51a0f6;
    }

    li a:hover {
      background-color: #2a2735;
    }

    li ul {
      list-style: none;
      padding-left: 0;
      margin: 0;
    }

    li a {
      color: #8c91a0;
      width: 150px;
      margin-left: 50px;
      line-height: 40px;
    }

    li a + ul {
      margin: 0 20px 0 15px;
    }
  }

}
</style>
