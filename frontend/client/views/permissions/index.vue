<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <tabs type="boxed" :is-fullwidth="false" alignment="centered" size="large">
        <tab-pane label="User Permissions" icon="fa fa-user">
          <manage-permissions :users="users" :groups="groups" :permission-options="permissionOptions"/>
        </tab-pane>
        <tab-pane label="User Groups" icon="fa fa-users">
          <manage-groups :groups="groups" :permission-options="permissionOptions"/>
        </tab-pane>
      </tabs>
    </div>
  </div>
</template>

<script>
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import {Collapse, Item as CollapseItem} from 'vue-bulma-collapse'
  import ManagePermissions from './permissions/manage-permissions'
  import ManageGroups from './permissions/manage-groups'
  import {EventBus} from '../../app'

  export default {
    components: {
      ManagePermissions,
      ManageGroups,
      Tabs,
      TabPane,
      Collapse,
      CollapseItem
    },

    data () {
      return {
        groups: [],
        users: [],
        permissionOptions: []
      }
    },

    mounted () {
      this.fetchData()
      EventBus.$on('refreshGroups', this.fetchData)
    },

    watch: {
      '$route': 'fetchData'
    },

    methods: {
      fetchData () {
        this.$http.get('/api/v1/permission')
          .then(response => {
            if (response.data) {
              this.permissionOptions = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.$http
          .get('/api/v1/permission/group', {showProgressBar: false})
          .then(response => {
            if (response.data) {
              this.groups = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.$http
          .get('/api/v1/users', {showProgressBar: false})
          .then(response => {
            if (response.data) {
              this.users = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
      },

      close () {
      }
    }
  }
</script>

<style lang="scss">

  .tabs {
    margin: 10px;

    .tab-content {
      min-height: 50px;
    }
  }

  .tabs.is-boxed li.is-active a {
    background-color: transparent;
    border-color: transparent;
    border-bottom-color: #4da2fc !important;
  }

</style>
