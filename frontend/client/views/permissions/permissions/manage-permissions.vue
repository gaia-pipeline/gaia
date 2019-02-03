<template>
  <div class="tile is-ancestor">
    <div class="tile is-horizontal is-parent is-3">
      <article class="tile is-child notification content-article box">
        <p class="control has-icons-left">
          <input v-model="search" class="input is-medium input-bar" type="text" placeholder="Search">
          <span class="icon is-small is-left"><i class="fa fa-search"></i></span>
        </p>
        <br>
        <table class="table is-narrow is-fullwidth table-general table-users">
          <thead>
          <tr>
            <th width="300">Username</th>
          </tr>
          </thead>
          <tbody>
          <tr v-if="filteredUsers.length > 0" v-for="user in filteredUsers" :key="user.username">
            <td class="user-row" @click="fetchData(user)">{{user.username}}</td>
          </tr>
          <tr v-if="filteredUsers.length === 0">
            <td class="user-row"><i>No results.</i></td>
          </tr>
          </tbody>
        </table>
      </article>
    </div>
    <div class="tile is-horizontal is-parent is-9">
      <article class="tile is-child notification content-article box">
        <div v-if="!permissions.username">
          <h4 class="title is-4">User Roles</h4>
          <p>Select a user from the list.</p>
        </div>
        <div v-else>
          <h4 class="title is-4">User Roles: {{ this.permissions.username }}</h4>
          <permission-tables :value="permissions.roles" @input="setRoles" :permission-options="permissionOptions"></permission-tables>
          <div style="float: left;">
            <button class="button is-primary" v-on:click="save">Save</button>
          </div>
        </div>
      </article>
    </div>
  </div>
</template>

<script>
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import {EventBus} from '../../../app'
  import PermissionTables from '../permissions/permission-tables'

  export default {
    name: 'manage-permissions',
    components: {Tabs, TabPane, PermissionTables},
    props: {
      reset: Function,
      users: Array,
      permissionOptions: Array
    },
    computed: {
      filteredUsers () {
        return this.users.filter(u => {
          return u.username.toLowerCase().includes(this.search.toLowerCase())
        })
      }
    },
    data () {
      return {
        search: '',
        permissions: {
          username: undefined,
          roles: []
        },
        permsString: ''
      }
    },
    mounted () {
      EventBus.$on('onUserDeleted', this.clear)
    },
    methods: {
      clear (username) {
        if (username === this.permissions.username) {
          this.permissions = {
            username: undefined,
            roles: []
          }
        }
      },
      setRoles (roles) {
        this.permissions.roles = roles
      },
      fetchData (user) {
        this.permissions = {
          username: user.username,
          roles: []
        }
        this.$http
          .get(`/api/v1/user/${user.username}/permissions`)
          .then(response => {
            if (response.data) {
              this.permissions = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
      },
      save () {
        this.$http
          .put(`/api/v1/user/${this.permissions.username}/permissions`, this.permissions)
          .then(() => this.$onSuccess('Permissions Updated', 'Permissions updated successfully.'))
          .catch((error) => this.$onError(error))
      }
    }
  }
</script>

<style scoped>
  .user-row {
    cursor: pointer;
  }

  .table-general {
    background: #413F4A;
    border: 2px solid #000;
  }

  .table-general th {
    border: 2px solid #000;
    background: #2c2b32;
    color: #4da2fc;
  }

  .table-general td {
    border: 2px solid #000;
    color: #8c91a0;
  }

  .table-users td:hover {
    border: 2px solid #000;
    background: #575463;
    cursor: pointer;
  }
</style>
