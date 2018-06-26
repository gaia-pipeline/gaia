<template>
  <section class="app-main" :style="[hiddenSidebarStyle]">
    <div class="container is-fluid is-marginless app-content">
      <headerbar></headerbar>
      <transition
        mode="out-in"
        enter-active-class="fadeIn"
        leave-active-class="fadeOut"
        appear>
        <router-view class="animated"></router-view>
      </transition>
    </div>
  </section>
</template>

<script>
import Headerbar from './Headerbar'
import { mapGetters } from 'vuex'

export default {
  computed: {
    ...mapGetters({
      sidebar: 'sidebar'
    }),
    hiddenSidebarStyle () {
      return this.sidebar.hidden ? { 'margin-left': 0 } : null
    }
  },

  components: {
    Headerbar
  }
}
</script>

<style lang="scss">
@import '~bulma/sass/utilities/mixins';

.app-main {
  padding-top: 0px;
  margin-left: 210px;

  @include mobile() {
    margin-left: 0;
  }

}

.app-content {
  padding: 20px;
  z-index: 0;
  padding-top: 90px;
}
</style>
