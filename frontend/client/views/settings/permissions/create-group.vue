<template>
  <modal :visible="visible" class="modal-z-index">
    <div class="box user-modal">
      <div class="block user-modal-content">
        <p class="control has-icons-left" style="padding-bottom: 5px;">
          <input class="input is-medium input-bar" v-focus type="text" v-model="group.name" placeholder="Name">
          <span class="icon is-small is-left">
                    <i class="fa fa-user"></i>
                  </span>
        </p>
        <div class="modal-footer">
          <div style="float: left;">
            <button class="button is-primary" v-on:click="add">Add</button>
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

  export default {
    components: {
      Modal
    },
    name: 'create-group',
    props: {
      visible: false
    },
    data () {
      return {
        permissions: [],
        group: {
          name: '',
          permissions: []
        }
      }
    },
    mounted () {
      this.getPermissionOptions()
    },
    methods: {
      getPermissionOptions () {
        this.$http
          .get('/api/v1/permission')
          .then(response => {
            if (response.data) {
              this.permissions = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
      },
      add () {
        this.$http
          .post('/api/v1/permission/group', this.group)
          .catch((error) => {
            this.$onError(error)
          })
        this.close()
      },
      close () {
        this.visible = false
      }
    }
  }
</script>

<style scoped>

</style>
