<template>
  <modal :visible="visible" class="modal-z-index">
    <div class="box user-modal">
      <collapse accordion is-fullwidth>
        <collapse-item :title="'Permissions: ' + user.display_name" selected>
          <div class="user-modal-content">
            <div v-for="po in permissionOptions">
              <input type="checkbox" :id="po" :value="po" v-model="permissions.permissions">
              <label :for="po">{{po}}</label>
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
              const perms = []
              response.data.forEach(pg => {
                pg.permissions.forEach(p => {
                  perms.push(pg.name + p.name)
                })
              })
              this.permissionOptions = perms
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
