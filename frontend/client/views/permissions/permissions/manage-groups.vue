<template>
  <div>
    <a class="button is-primary" v-on:click="toggleNew(true)">
      <span class="icon">
        <i class="fa fa-user-plus"></i>
      </span>
      <span>Add Group</span>
    </a><br><br>
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
              <th width="300">Groups</th>
            </tr>
            </thead>
            <tbody>
            <tr v-if="filteredGroups.length > 0" v-for="group in filteredGroups" :key="group.name">
              <td class="user-row" @click="selectGroup(group)">{{group.name}}</td>
            </tr>
            <tr v-if="filteredGroups.length === 0">
              <td class="user-row"><i>No results.</i></td>
            </tr>
            </tbody>
          </table>
        </article>
      </div>
      <div class="tile is-horizontal is-parent is-9">
        <article class="tile is-child notification content-article box">
          <div v-if="isNew">
            <h4 class="title is-4">User Groups: New</h4>
            <input class="input is-medium input-bar" v-focus v-model="name" type="text" placeholder="Name">
            <br><br>
            <input class="input is-medium input-bar" v-focus v-model="description" type="text"
                   placeholder="Description">
            <br><br>
            <permission-tables @input="setRoles" :permission-options="permissionOptions"></permission-tables>
            <div style="float: left;">
              <button class="button is-primary" v-on:click="addNew">Add</button>
            </div>
          </div>
          <div v-if="!isNew && name !== ''">
            <h4 class="title is-4">User Groups: {{name}}</h4>
            <h4 class="title is-5">{{description ? description : "No description"}}</h4>
            <permission-tables :value="roles" @input="setRoles" :permission-options="permissionOptions"></permission-tables>
            <div style="float: left;">
              <button class="button is-primary" v-on:click="save">Save</button>
            </div>
          </div>
          <div v-if="!isNew && name === ''">
            <h4 class="title">User Groups</h4>
            <p>Please select an existing group or create a new one.</p>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import PermissionTables from './permission-tables'
  import {EventBus} from '../../../app'

  export default {
    name: 'manage-groups',
    components: {PermissionTables, Tabs, TabPane},
    props: {
      reset: Function,
      groups: Array,
      permissionOptions: Array
    },
    computed: {
      filteredGroups () {
        return this.groups.filter(g => {
          return g.name.toLowerCase().includes(this.search.toLowerCase())
        })
      }
    },
    data () {
      return {
        search: '',
        isNew: false,
        roles: [],
        name: '',
        description: ''
      }
    },
    methods: {
      setRoles (roles) {
        this.roles = roles
      },
      selectGroup (group) {
        this.toggleNew(false)
        this.name = group.name
        this.description = group.description
        this.roles = group.roles
      },
      toggleNew (value) {
        this.name = ''
        this.description = ''
        this.roles = []
        this.isNew = value
      },
      addNew () {
        const perms = {
          name: this.name,
          roles: this.roles,
          description: this.description
        }
        this.$http
          .post(`/api/v1/permission/group`, perms)
          .then(() => {
            this.$onSuccess('Group Added', 'Permission group added successfully.')
            EventBus.$emit('refreshGroups')
          })
          .catch((error) => this.$onError(error))
      },
      save () {
        const perms = {
          name: this.name,
          roles: this.roles,
          description: this.description
        }
        this.$http
          .put(`/api/v1/permission/group`, perms)
          .then(() => {
            this.$onSuccess('Group Updated', 'Permission group updated successfully.')
            EventBus.$emit('refreshGroups')
          })
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
