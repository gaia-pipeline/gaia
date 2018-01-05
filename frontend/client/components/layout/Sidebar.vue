<template>
  <aside class="menu app-sidebar animated" :class="{ slideInLeft: show, slideOutLeft: !show }">
    <div>
      <a class="navbar-item brand-top" href="/">
        <img src="~assets/logo.png">
        &nbsp;&nbsp;<div class="header-text">Gaia</div>
      </a>
    </div>
    <ul class="menu-list">
      <li v-for="(item, index) in menu">
        <router-link :to="item.path" :exact="true" :aria-expanded="isExpanded(item) ? 'true' : 'false'" v-if="item.path" @click.native="toggle(index, item)">
          <span class="icon icon-left is-small"><i :class="['fa', item.meta.icon]"></i></span>
          {{ item.meta.label || item.name }}
          <span class="icon is-small is-angle" v-if="item.children && item.children.length">
            <i class="fa fa-angle-down"></i>
          </span>
        </router-link>
        <a :aria-expanded="isExpanded(item)" v-else @click="toggle(index, item)">
          <span class="icon is-small"><i :class="['fa', item.meta.icon]"></i></span>
          {{ item.meta.label || item.name }}
          <span class="icon is-small is-angle" v-if="item.children && item.children.length">
            <i class="fa fa-angle-down"></i>
          </span>
        </a>

        <expanding v-if="item.children && item.children.length">
          <ul v-show="isExpanded(item)">
            <li v-for="subItem in item.children" v-if="subItem.path">
              <router-link :to="generatePath(item, subItem)">
                {{ subItem.meta && subItem.meta.label || subItem.name }}
              </router-link>
            </li>
          </ul>
        </expanding>
      </li>
    </ul>
  </aside>
</template>

<script>
import Expanding from 'vue-bulma-expanding'
import { mapGetters, mapActions } from 'vuex'

export default {
  components: {
    Expanding
  },

  props: {
    show: Boolean
  },

  data () {
    return {
      isReady: false
    }
  },

  mounted () {
    let route = this.$route
    if (route.name) {
      this.isReady = true
      this.shouldExpandMatchItem(route)
    }
  },

  computed: mapGetters({
    menu: 'menuitems'
  }),

  methods: {
    ...mapActions([
      'expandMenu'
    ]),

    isExpanded (item) {
      return item.meta.expanded
    },

    toggle (index, item) {
      this.expandMenu({
        index: index,
        expanded: !item.meta.expanded
      })
    },

    shouldExpandMatchItem (route) {
      let matched = route.matched
      let lastMatched = matched[matched.length - 1]
      let parent = lastMatched.parent || lastMatched
      const isParent = parent === lastMatched

      if (isParent) {
        const p = this.findParentFromMenu(route)
        if (p) {
          parent = p
        }
      }

      if ('expanded' in parent.meta && !isParent) {
        this.expandMenu({
          item: parent,
          expanded: true
        })
      }
    },

    generatePath (item, subItem) {
      return `${item.component ? item.path + '/' : ''}${subItem.path}`
    },

    findParentFromMenu (route) {
      const menu = this.menu
      for (let i = 0, l = menu.length; i < l; i++) {
        const item = menu[i]
        const k = item.children && item.children.length
        if (k) {
          for (let j = 0; j < k; j++) {
            if (item.children[j].name === route.name) {
              return item
            }
          }
        }
      }
    }
  },

  watch: {
    $route (route) {
      this.isReady = true
      this.shouldExpandMatchItem(route)
    }
  }

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
    &.is-angle {
      position: absolute;
      right: 10px;
      margin-top: 13px;
      transition: transform .377s ease;
    }
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

  .menu-list {  
    margin: auto;
    width: 200px;

    li {
      float: left;
    }

    li a.is-active {
      background-color: rgb(60, 57, 74);
      color: #51a0f6;
      font-weight: bold;
      border-right: 3px solid #51a0f6;
    }

    li a:hover {
      background-color: #2a2735;
    }

    li a {
      color: #8c91a0;
      width: 140px;
      margin-left: 60px;
      line-height: 40px;

      &[aria-expanded="true"] {
        .is-angle {
          transform: rotate(180deg);
        }
      }
    }

    li a + ul {
      margin: 0 20px 0 15px;
    }
  }

}
</style>
