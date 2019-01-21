<template>
  <modal :visible="visible" class="modal-z-index">
    <div class="box user-modal">
      <collapse accordion is-fullwidth>
        <collapse-item :title="'Roles: ' + user.username" selected>
          <collapse accordion is-fullwidth>
            <collapse-item v-for="category in permissionOptions" :key="category.name" :title="category.name" selected>
              <div class="user-modal-content">
                <a class="button is-primary is-small" v-on:click="selectAll(category)">
                  <span>Select All</span>
                </a>
                <a class="button is-primary is-small" v-on:click="deselectAll(category)">
                  <span>Deselect All</span>
                </a>
                <br><br>
                  <div v-for="role in category.roles">
                    <input type="checkbox" :id="getFullName(category, role)" :value="getFullName(category, role)" v-model="permissions.roles">
                    <label :for="getFullName(category, role)">{{role.name}}</label>
                  </div>
              </div>
            </collapse-item>
          </collapse>
        </collapse-item>
      </collapse>
      <div class="block user-modal-content">
        <div class="modal-footer">
          <div style="float: left;">
            <button class="button is-primary" v-on:click="save">Save</button>
          </div>
          <div style="float: right;">
            <button class="button is-danger" v-on:click="close">Cancel</button>
          </div>
        </div>
      </div>
    </div>
  </modal>
</template>

<script>
  import {Modal} from 'vue-bulma-modal'
  import {Collapse, Item as CollapseItem} from 'vue-bulma-collapse'

  export default {
    name: 'user-permissions',
    components: {Modal, CollapseItem, Collapse},
    props: {
      visible: Boolean,
      user: Object
    },
    data () {
      return {
        permissions: {
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
      flattenOptions (category) {
        return category.roles.map(p => category.name + p.name)
      },
      getFullName (category, role) {
        return category.name + role.name
      },
      fetchData () {
        this.$http
          .get(`/api/v1/user/${this.user.username}/permissions`)
          .then(response => {
            if (response.data) {
              this.permissions = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.$http
          .get('/api/v1/permission')
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
        this.permissions.username = this.user.username
        this.$http
          .put(`/api/v1/user/${this.user.display_name}/permissions`, this.permissions)
          .then(() => {
            this.close()
          })
          .catch((error) => {
            this.$onError(error)
          })
      },
      close () {
        this.permissions = {
          roles: []
        }
        this.$emit('close')
      }
    }
  }
</script>

<style scoped>

</style>
