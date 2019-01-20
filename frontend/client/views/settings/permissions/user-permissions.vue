<template>
  <modal :visible="visible" class="modal-z-index">
    <div class="box user-modal">
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
              <div v-for="permission in category.permissions">
                <input type="checkbox" :id="getFullName(category, permission)" :value="getFullName(category, permission)" v-model="permissions.permissions">
                <label :for="getFullName(category, permission)">{{permission.name}}</label>
              </div>
          </div>
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
          permissions: []
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
          if (this.permissions.permissions.indexOf(p) === -1) {
            this.permissions.permissions.push(p)
          }
        })
      },
      deselectAll (category) {
        this.flattenOptions(category).forEach(p => {
          let index = this.permissions.permissions.indexOf(p)
          if (index > -1) {
            this.permissions.permissions.splice(index, 1)
          }
        })
      },
      flattenOptions (category) {
        return category.permissions.map(p => category.name + p.name)
      },
      getFullName (category, permission) {
        return category.name + permission.name
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
          permissions: []
        }
        this.$emit('close')
      }
    }
  }
</script>

<style scoped>

</style>
