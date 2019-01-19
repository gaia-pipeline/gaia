<template>
  <div class="tile is-ancestor">
    <create-group :visible="showCreateDialog"></create-group>
    <div class="tile is-vertical">
      <div class="tile is-parent">
        <a class="button is-primary" v-on:click="createGroup" style="margin-bottom: -10px;">
                  <span class="icon">
                    <i class="fa fa-plus"></i>
                  </span>
          <span>Create Group</span>
        </a>
      </div>
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <vue-good-table
            :columns="columns"
            :rows="groups"
            :paginate="true"
            :global-search="true"
            globalSearchPlaceholder="Search ..."
            styleClass="table table-own-bordered">
            <template slot="table-row" slot-scope="props">
              <td>
                <span>{{ props.row.name }}</span>
              </td>
              <td>
                <span>{{ formatPermissions(props.row.permissions) }}</span>
              </td>
              <td>
                <a v-on:click="editGroup(props.row)"><i class="fa fa-edit" style="color: whitesmoke;"></i></a>
                <a v-on:click="deleteGroup(props.row)"><i class="fa fa-delete" style="color: whitesmoke;"></i></a>
              </td>
            </template>
            <div slot="emptystate" class="empty-table-text">
              No permission groups.
            </div>
          </vue-good-table>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
  import Assign from './assign'
  import CreateGroup from './create-group'

export default {
    name: 'manage',
    components: {CreateGroup, Assign},
    data () {
      return {
        showCreateDialog: false,
        groups: [],
        columns: [
          {
            label: 'Name',
            field: 'name'
          },
          {
            label: 'Permissions',
            field: 'permissions'
          },
          {
            label: '',
            field: 'actions'
          }
        ]
      }
    },
    mounted () {
      this.fetchData()
    },
    methods: {
      fetchData () {
        this.$http
          .get('/api/v1/permission/group')
          .then(response => {
            if (response.data) {
              this.groups = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
      },
      createGroup () {
        this.showCreateDialog = true
      },
      editGroup (value) {
      },
      deleteGroup (value) {
      },
      formatPermissions (perms) {
        return perms.join(', ')
      }
    }
  }
</script>

<style scoped>

</style>
