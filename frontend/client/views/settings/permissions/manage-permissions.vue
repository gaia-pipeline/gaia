<template>
  <div class="tile is-ancestor">
    <div class="tile is-horizontal is-parent is-3">
      <article class="tile is-child notification content-article box">
        <p class="control has-icons-left">
          <input v-model="search" class="input is-medium input-bar" type="text" placeholder="Search">
          <span class="icon is-small is-left"><i class="fa fa-search"></i></span>
        </p>
        <br>
        <table class="table is-striped is-hoverable is-fullwidth">
          <thead>
          <tr>
            <th width="300">Username</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="user in filteredUsers" :key="user.username">
            <td class="user-row" @click="fetchData(user)">{{user.username}}</td>
          </tr>
          </tbody>
        </table>
      </article>
    </div>
    <div class="tile is-horizontal is-parent is-9">
      <article class="tile is-child notification content-article box">
        <div v-if="!this.permissions.username">
          <h4 class="title is-4">Permission Roles: None</h4>
          <p>Select a user from the list.</p>
        </div>
        <div v-else>
          <h4 class="title is-4">Permission Roles: {{ this.permissions.username }}</h4>
          <div v-for="category in permissionOptions" :key="category.name">
            <p>{{ category.name }}: {{ category.description }}</p><br>
            <table class="table is-striped is-narrow is-hoverable is-fullwidth">
              <thead>
              <tr>
                <th width="60"><input type="checkbox" @click="checkAll(category)" :checked="allSelected(category)"/>
                </th>
                <th width="300">Name</th>
                <th>Description</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="role in category.roles">
                <td><input type="checkbox" :id="getFullName(category, role)" :value="getFullName(category, role)"
                           v-model="permissions.roles"></td>
                <td>{{role.name}}</td>
                <td>Desc...</td>
              </tr>
              </tbody>
            </table>
          </div>
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

  export default {
    name: 'manage-permissions',
    components: {Tabs, TabPane},
    props: {
      users: Array
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
        permsString: '',
        permissionOptions: []
      }
    },
    watch: {
      visible (newVal) {
        if (newVal) {
          this.fetchData()
        }
      }
    },
    methods: {
      checkAll (category) {
        if (this.allSelected(category)) {
          this.deselectAll(category)
        } else {
          this.selectAll(category)
        }
      },
      selectAll (category) {
        this.flattenOptions(category).forEach(p => {
          if (this.permissions.roles.indexOf(p) === -1) {
            this.permissions.roles.push(p)
          }
        })
      },
      deselectAll (category) {
        this.flattenOptions(category).forEach(p => {
          let index = this.permissions.roles.indexOf(p)
          if (index > -1) {
            this.permissions.roles.splice(index, 1)
          }
        })
      },
      allSelected (category) {
        for (let role of category.roles) {
          const name = this.getFullName(category, role)
          if (this.permissions.roles.indexOf(name) === -1) {
            return false
          }
        }
        return true
      },
      flattenOptions (category) {
        return category.roles.map(p => category.name + p.name)
      },
      getFullName (category, role) {
        return category.name + role.name
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
            return this.$http.get('/api/v1/permission')
          })
          .then(response => {
            if (response.data) {
              this.permissionOptions = response.data
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
</style>
