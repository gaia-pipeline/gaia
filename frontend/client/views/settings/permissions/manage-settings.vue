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
            <th width="300">Settings</th>
          </tr>
          </thead>
          <tbody>
            <td class="settings-row"><i>Poller</i></td>
          </tbody>
        </table>
      </article>
    </div>
  </div>
</template>

<script>
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import {EventBus} from '../../../app'
  import VbSwitch from 'vue-bulma-switch'

  export default {
    name: 'manage-settings',
    components: {Tabs, TabPane, VbSwitch},
    props: {
      reset: Function,
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
        permissionOptions: [],
        settingsTogglePoller: false,
        settingsTogglePollerText: 'Off'
      }
    },
    methods: {
      settingsTogglePollerSwitch(val) {
        // TODO: Get and Send to API here.
        if (val) {
          this.settingsTogglePollerText = 'On'
        } else {
          this.settingsTogglePollerText = 'Off'
        }
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
