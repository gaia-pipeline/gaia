<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <tabs type="boxed" :is-fullwidth="true" alignment="centered" size="large">
        <tab-pane label="Manage Users" icon="fa fa-user-circle">
          <div class="tile is-parent">
            <article class="tile is-child notification content-article box">
              <vue-good-table
                :columns="userColumns"
                :rows="userRows"
                :paginate="true"
                :global-search="true"
                :defaultSortBy="{field: 'username', type: 'desc'}"
                globalSearchPlaceholder="Search ..."
                styleClass="table table-own-bordered">
                <template slot="table-row" slot-scope="props">
                  <td>
                    <span>{{ props.row.display_name }}</span>
                  </td>
                  <td :title="props.row.lastlogin" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                    <span>{{ convertTime(props.row.lastlogin) }}</span>
                  </td>
                  <td>
                    <a v-on:click="editUserModal(props.row)"><i class="fa fa-edit" style="color: whitesmoke;"></i></a>
                    <a v-on:click="deleteUserModal(props.row)"><i class="fa fa-trash" style="color: whitesmoke;"></i></a>
                  </td>
                </template>
                <div slot="emptystate" class="empty-table-text">
                  No users found in database.
                </div>
              </vue-good-table>
            </article>
          </div>
        </tab-pane>
        <!--<tab-pane label="Manage Pipelines" icon="fa fa-wrench"></tab-pane>-->
      </tabs>
    </div>

    <!-- edit user modal -->
    <modal :visible="showEditUserModal" class="modal-z-index" @close="close">
      <div class="box user-modal">
        <div class="block edit-user-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Change Password" selected>
              <div class="edit-user-modal-content">
                <label class="label" style="text-align: left;">Change password for user {{ selectUser.display_name }}:</label>
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="password" v-model="selectUser.oldpassword" placeholder="Old Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.newpassword" placeholder="New Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.newpasswordconf" placeholder="New Password confirmation">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="changePassword">Change Password</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="cancel">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- delete user modal -->
    <modal :visible="showDeleteUserModal" class="modal-z-index" @close="close">
      <div class="box user-modal">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <p>
                <span style="color: whitesmoke;">Do you really want to delete the user {{ selectUser.display_name }}?</span>
              </p>
            </div>
            <div class="modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="deleteUser" style="width:150px;">Yes</button>
              </div>
              <div style="float: right;">
                <button class="button is-danger" v-on:click="cancel" style="width:130px;">No</button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </modal>
  </div>
</template>

<script>
import Vue from 'vue'
import { Tabs, TabPane } from 'vue-bulma-tabs'
import { Modal } from 'vue-bulma-modal'
import { Collapse, Item as CollapseItem } from 'vue-bulma-collapse'
import VueGoodTable from 'vue-good-table'
import VueTippy from 'vue-tippy'
import moment from 'moment'
import Notification from 'vue-bulma-notification-fixed'

const NotificationComponent = Vue.extend(Notification)
const openNotification = (propsData = {
  title: '',
  message: '',
  type: '',
  direction: '',
  duration: 4500,
  container: '.notifications'
}) => {
  return new NotificationComponent({
    el: document.createElement('div'),
    propsData
  })
}

Vue.use(VueGoodTable)
Vue.use(VueTippy)

export default {
  components: {
    Tabs,
    TabPane,
    Modal,
    Collapse,
    CollapseItem
  },

  data () {
    return {
      userColumns: [
        {
          label: 'Name',
          field: 'display_name'
        },
        {
          label: 'Last Login',
          field: 'lastlogin'
        },
        {
          label: ''
        }
      ],
      userRows: [],
      selectUser: [],
      showEditUserModal: false,
      showDeleteUserModal: false
    }
  },

  mounted () {
    this.fetchData()
  },

  watch: {
    '$route': 'fetchData'
  },

  methods: {
    fetchData () {
      this.$http
        .get('/api/v1/users', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.userRows = response.data
          }
        })
        .catch((error) => {
          this.$onError(error)
        })
    },

    convertTime (time) {
      return moment(time).fromNow()
    },

    editUserModal (user) {
      this.selectUser = user
      this.showEditUserModal = true
    },

    deleteUserModal (user) {
      this.selectUser = user
      this.showDeleteUserModal = true
    },

    close () {
      this.showEditUserModal = false
      this.showDeleteUserModal = false
      this.$emit('close')
    },

    cancel () {
      // cancel means reset all stuff
      this.selectUser.oldpassword = ''
      this.selectUser.newpassword = ''
      this.selectUser.newpasswordconf = ''

      this.close()
    },

    changePassword () {
      // pre-validate
      if (this.selectUser.newpassword === '' || this.selectUser.newpasswordconf === '') {
        openNotification({
          title: 'Empty password',
          message: 'Empty password is not allowed.',
          type: 'danger'
        })
        this.close()
        return
      }

      this.$http
        .post('/api/v1/user/password', this.selectUser)
        .then(response => {
          openNotification({
            title: 'Password changed!',
            message: 'Password has been successful changed.',
            type: 'success'
          })
        })
        .catch((error) => {
          this.$onError(error)
        })
      this.close()
    }
  },

  deleteUser () {
    this.$http
      .delete('/api/v1/user', this.selectUser)
      .then(response => {
        openNotification({
          title: 'User deleted!',
          message: 'User ' + this.selectUser.display_name + ' has been successfully deleted.',
          type: 'success'
        })
      })
      .catch((error) => {
        this.$onError(error)
      })
    this.close()
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

.user-modal {
  text-align: center;
  background-color: #2a2735;
}

.edit-user-modal-content {
  margin: auto;
  padding: 10px;
}

.modal-footer {
  height: 45px;
  padding-top: 15px;
}

</style>
