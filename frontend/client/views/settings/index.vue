<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <tabs type="boxed" :is-fullwidth="false" alignment="centered" size="large">
        <tab-pane label="Users" icon="fa fa-user-circle">
          <div class="tile is-ancestor">
            <div class="tile is-vertical">
              <div class="tile is-parent">
                <a class="button is-primary" v-on:click="addUserModal" style="margin-bottom: -10px;">
                  <span class="icon">
                    <i class="fa fa-user-plus"></i>
                  </span>
                  <span>Create User</span>
                </a>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification content-article box">
                  <vue-good-table
                    :columns="userColumns"
                    :rows="userRows"
                    :pagination-options="{
                      enabled: true,
                      mode: 'records'
                    }"
                    :search-options="{enabled: true, placeholder: 'Search ...'}"
                    :sort-options="{
                      enabled: true,
                      initialSortBy: {field: 'display_name', type: 'desc'}
                    }"
                    styleClass="table table-grid table-own-bordered">
                    <template slot="table-row" slot-scope="props">
                      <span v-if="props.column.field === 'display_name'">{{ props.row.display_name }}</span>
                      <span v-if="props.column.field === 'lastlogin'":title="props.row.lastlogin" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                        {{ convertTime(props.row.lastlogin) }}
                      </span>
                      <span v-if="props.column.field === 'trigger_token'" :title="props.row.trigger_token" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                        {{ props.row.trigger_token }}
                      </span>
                      <span v-if="props.column.field === 'action'">
                        <a v-on:click="editUserModal(props.row)"><i class="fa fa-edit"
                                                                    style="color: whitesmoke;"></i></a>
                        <a v-on:click="resetTriggerTokenModal(props.row)" v-if="props.row.username === 'auto'">
                                                                <i class="fa fa-sliders" style="color: whitesmoke;"></i></a>
                        <a v-on:click="deleteUserModal(props.row)" v-if="props.row.username !== session.username && props.row.username !== 'auto'"><i
                          class="fa fa-trash" style="color: whitesmoke;"></i></a>
                      </span>
                    </template>
                    <div slot="emptystate" class="empty-table-text">
                      No users found in database.
                    </div>
                  </vue-good-table>
                </article>
              </div>
            </div>
          </div>
        </tab-pane>
        <tab-pane label="Permissions" icon="fa fa-users">
          <manage-permissions :users="userRows"/>
        </tab-pane>
        <tab-pane label="Pipelines" icon="fa fa-cog">
          <div class="tile is-ancestor">
            <div class="tile is-vertical">
              <div class="tile is-parent">
                <a class="button is-primary" v-on:click="createPipeline" style="margin-bottom: -10px;">
                  <span class="icon">
                    <i class="fa fa-plus"></i>
                  </span>
                  <span>Create Pipeline</span>
                </a>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification content-article box">
                  <vue-good-table
                    :columns="pipelineColumns"
                    :rows="pipelineRows"
                    :pagination-options="{
                      enabled: true,
                      mode: 'records'
                    }"
                    :search-options="{enabled: true, placeholder: 'Search ...'}"
                    :sort-options="{
                      enabled: true,
                      initialSortBy: {field: 'id', type: 'desc'}
                    }"
                    styleClass="table table-grid table-own-bordered">
                    <template slot="table-row" slot-scope="props">
                      <span v-if="props.column.field === 'name'">
                        <span>{{ props.row.name }}</span>
                      </span>
                      <span v-if="props.column.field === 'type'">
                        <span>{{ props.row.type }}</span>
                      </span>
                      <span v-if="props.column.field === 'created'">
                        <span>{{ convertTime(props.row.created) }}</span>
                      </span>
                      <span v-if="props.column.field === 'action'">
                        <a v-on:click="editPipelineModal(props.row)"><i class="fa fa-edit"
                                                                        style="color: whitesmoke;"></i></a>
                        <a v-on:click="resetPipelineTriggerTokenModal(props.row)"><i class="fa fa-sliders"
                                                                        style="color: whitesmoke;"></i></a>
                        <a v-on:click="deletePipelineModal(props.row)"><i class="fa fa-trash"
                                                                          style="color: whitesmoke;"></i></a>
                      </span>
                    </template>
                    <div slot="emptystate" class="empty-table-text">
                      No active pipelines.
                    </div>
                  </vue-good-table>
                </article>
              </div>
            </div>
          </div>
        </tab-pane>
        <tab-pane label="Worker" icon="fa fa-user-secret">
          <manage-worker/>
        </tab-pane>
        <tab-pane label="Settings" icon="fa fa-wrench">
          <manage-settings/>
        </tab-pane>
      </tabs>
    </div>

    <!-- edit user modal -->
    <modal :visible="showEditUserModal" class="modal-z-index" @close="close">
      <div class="box user-modal">
        <div class="block user-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Change Password" selected>
              <div class="user-modal-content">
                <label class="label" style="text-align: left;">Change password for user {{ selectUser.display_name
                  }}:</label>
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="password" v-model="selectUser.oldpassword"
                         placeholder="Old Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.newpassword"
                         placeholder="New Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.newpasswordconf"
                         placeholder="New Password confirmation">
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
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- reset trigger token modal -->
    <modal :visible="showResetTriggerTokenModal" class="modal-z-index" @close="close">
      <div class="box user-modal">
        <div class="block user-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Reset Trigger Token" selected>
              <div class="user-modal-content">
                <label class="label" style="text-align: left;">Reset Trigger Token for user {{ selectUser.display_name
                  }}?</label>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="resetUserTriggerToken">Reset Token</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
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
                <span
                  style="color: whitesmoke;">Do you really want to delete the user {{ selectUser.display_name }}?</span>
              </p>
            </div>
            <div class="modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="deleteUser" style="width:150px;">Yes</button>
              </div>
              <div style="float: right;">
                <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </modal>

    <!-- add user modal -->
    <modal :visible="showAddUserModal" class="modal-z-index" @close="close">
      <div class="box user-modal">
        <div class="block user-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Add User" selected>
              <div class="user-modal-content">
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="text" v-model="selectUser.username"
                         placeholder="Username">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="text" v-model="selectUser.display_name"
                         placeholder="Display Name (optional)">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user-secret"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.password"
                         placeholder="Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectUser.passwordconf"
                         placeholder="Password confirmation">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="addUser">Add User</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- edit pipeline modal -->
    <modal :visible="showEditPipelineModal" class="modal-z-index" @close="close">
      <div class="box pipeline-modal">
        <div class="block pipeline-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Change Pipeline Name" selected>
              <div class="pipeline-modal-content">
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus v-model="selectPipeline.name"
                         placeholder="Pipeline Name">
                  <span class="icon is-small is-left">
                    <i class="fa fa-book"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
            <collapse-item title="Change Periodic Schedule">
              <div class="pipeline-modal-content">
                <p class="control">
                  <textarea class="textarea input-bar" v-model="pipelinePeriodicSchedules"></textarea>
                </p>
                <label class="label" style="test-align: left;">
                  Use the standard cron syntax. For example to start the pipeline every half hour:
                  <br/>0 30 * * * *<br/>
                  Please see <a href="https://godoc.org/github.com/robfig/cron" target="_blank">here</a> for more
                  information.
                </label>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="changePipeline">Accept changes</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- reset trigger token modal -->
   <modal :visible="showResetPipelineTriggerTokenModal" class="modal-z-index" @close="close">
      <div class="box pipeline-modal">
        <div class="block pipeline-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Reset Pipeline Trigger Token" selected>
              <div class="pipeline-modal-content">
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <label class="label" style="text-align: left;">Reset Token for pipeline {{ selectPipeline.name
                    }}?</label>
                </p>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="resetPipelineTriggerToken">Reset Token</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- delete pipeline modal -->
    <modal :visible="showDeletePipelineModal" class="modal-z-index" @close="close">
      <div class="box pipeline-modal">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <p>
                <span
                  style="color: whitesmoke;">Do you really want to delete the pipeline "{{ selectPipeline.name }}"?</span>
              </p>
            </div>
            <div class="modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="deletePipeline" style="width:150px;">Yes</button>
              </div>
              <div style="float: right;">
                <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
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
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import {Modal} from 'vue-bulma-modal'
  import {Collapse, Item as CollapseItem} from 'vue-bulma-collapse-fixed'
  import { VueGoodTable } from 'vue-good-table'
  import 'vue-good-table/dist/vue-good-table.css'
  import VueTippy from 'vue-tippy'
  import moment from 'moment'
  import Notification from 'vue-bulma-notification-fixed'
  import {mapGetters} from 'vuex'
  import ManagePermissions from './permissions/manage-permissions'
  import ManageSettings from './settings/manage-settings'
  import ManageWorker from './worker/manage-worker'
  import {EventBus} from '../../app'

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

  Vue.use(VueTippy)

  export default {
    components: {
      ManagePermissions,
      Tabs,
      TabPane,
      Modal,
      Collapse,
      CollapseItem,
      ManageSettings,
      ManageWorker,
      VueGoodTable
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
            label: 'Trigger Token',
            field: 'trigger_token'
          },
          {
            label: 'Action',
            field: 'action'
          }
        ],
        userRows: [],
        pipelineColumns: [
          {
            label: 'Name',
            field: 'name'
          },
          {
            label: 'Type',
            field: 'type'
          },
          {
            label: 'Created',
            field: 'created'
          },
          {
            label: 'Action',
            field: 'action'
          }
        ],
        pipelineRows: [],
        selectUser: {},
        selectPipeline: {},
        showEditUserModal: false,
        showDeleteUserModal: false,
        showResetTriggerTokenModal: false,
        showAddUserModal: false,
        showEditPipelineModal: false,
        showResetPipelineTriggerTokenModal: false,
        showDeletePipelineModal: false,
        pipelinePeriodicSchedules: ''
      }
    },

    mounted () {
      this.fetchData()
    },

    watch: {
      '$route': 'fetchData'
    },

    computed: mapGetters({
      session: 'session'
    }),

    methods: {
      fetchData () {
        this.$http
          .get('/api/v1/users', { params: { hideProgressBar: true }})
          .then(response => {
            if (response.data) {
              this.userRows = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.$http
          .get('/api/v1/pipeline', { params: { hideProgressBar: true }})
          .then(response => {
            if (response.data) {
              this.pipelineRows = response.data
            } else {
              this.pipelineRows = []
            }
          }).catch((error) => {
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

      resetTriggerTokenModal (user) {
        this.selectUser = user
        this.showResetTriggerTokenModal = true
      },

      deleteUserModal (user) {
        this.selectUser = user
        this.showDeleteUserModal = true
      },

      addUserModal () {
        this.selectUser = {}
        this.showAddUserModal = true
      },

      editPipelineModal (pipeline) {
        // Check if periodic schedules is given.
        if (pipeline.periodicschedules) {
          this.pipelinePeriodicSchedules = pipeline.periodicschedules.join('\n')
        }

        this.selectPipeline = pipeline
        this.showEditPipelineModal = true
      },

      resetPipelineTriggerTokenModal (pipeline) {
        this.selectPipeline = pipeline
        this.showResetPipelineTriggerTokenModal = true
      },

      deletePipelineModal (pipeline) {
        this.selectPipeline = pipeline
        this.showDeletePipelineModal = true
      },

      close () {
        this.showEditUserModal = false
        this.showDeleteUserModal = false
        this.showResetTriggerTokenModal = false
        this.showAddUserModal = false
        this.selectUser = {}
        this.showEditPipelineModal = false
        this.showResetPipelineTriggerTokenModal = false
        this.showDeletePipelineModal = false
        this.selectPipeline = {}
        this.pipelinePeriodicSchedules = ''
        this.$emit('close')
      },

      changePassword () {
        // pre-validate
        if (!this.selectUser.newpassword || !this.selectUser.newpasswordconf) {
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
      },

      resetUserTriggerToken () {
        this.$http
          .put('/api/v1/user/' + this.selectUser.username + '/reset-trigger-token')
          .then(response => {
            openNotification({
              title: 'Token changed!',
              message: 'New trigger token has been generated!',
              type: 'success'
            })
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.close()
      },

      addUser () {
        // pre-validate
        if (!this.selectUser.password || !this.selectUser.passwordconf) {
          openNotification({
            title: 'Empty password',
            message: 'Empty password is not allowed.',
            type: 'danger'
          })
          this.close()
          return
        }

        // pre-validate
        if (!this.selectUser.username || this.selectUser.username.trim() === '') {
          openNotification({
            title: 'Empty username',
            message: 'Empty username is not allowed.',
            type: 'danger'
          })
          this.close()
          return
        }

        // pre-validate
        if (this.selectUser.password !== this.selectUser.passwordconf) {
          openNotification({
            title: 'Password not identical',
            message: 'Password and confirmation are not identical!',
            type: 'danger'
          })
          this.close()
          return
        }
        this.selectUser.passwordconf = null

        // Display name is optional
        if (!this.selectUser.display_name) {
          this.selectUser.display_name = this.selectUser.username
        }

        this.$http
          .post('/api/v1/user', this.selectUser)
          .then(response => {
            openNotification({
              title: 'User added!',
              message: 'User has been successfully added.',
              type: 'success'
            })
            this.fetchData()
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.close()
      },

      deleteUser () {
        this.$http
          .delete('/api/v1/user/' + this.selectUser.username)
          .then(response => {
            openNotification({
              title: 'User deleted!',
              message: 'User ' + this.selectUser.display_name + ' has been successfully deleted.',
              type: 'success'
            })
            EventBus.$emit('onUserDeleted', this.selectUser.username)
            this.fetchData()
            this.close()
          })
          .catch((error) => {
            this.$onError(error)
          })
      },

      createPipeline () {
        this.$router.push('/pipeline/create')
      },

      changePipeline () {
        // Convert periodic schedules into list.
        this.selectPipeline.periodicschedules = this.pipelinePeriodicSchedules.split('\n')

        this.$http
          .put('/api/v1/pipeline/' + this.selectPipeline.id, this.selectPipeline)
          .then(response => {
            openNotification({
              title: 'Pipeline updated!',
              message: 'Pipeline has been successfully updated.',
              type: 'success'
            })
            this.fetchData()
            this.close()
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.close()
      },

      resetPipelineTriggerToken () {
        this.$http
          .put('/api/v1/pipeline/' + this.selectPipeline.id + '/reset-trigger-token')
          .then(response => {
            openNotification({
              title: 'Token changed!',
              message: 'New trigger token has been generated!',
              type: 'success'
            })
          })
          .catch((error) => {
            this.$onError(error)
          })
        this.close()
      },

      deletePipeline () {
        this.$http
          .delete('/api/v1/pipeline/' + this.selectPipeline.id)
          .then(response => {
            openNotification({
              title: 'Pipeline deleted!',
              message: 'Pipeline ' + this.selectPipeline.name + ' has been successfully deleted.',
              type: 'success'
            })
            this.fetchData()
            this.close()
          })
          .catch((error) => {
            this.$onError(error)
          })
      }
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

  .tabs.is-boxed a:hover {
    background-color: black;
    color: #4da2fc;
    border-bottom-color: #4da2fc;
  }

  .user-modal, .pipeline-modal {
    text-align: center;
    background-color: #2a2735;
  }

  .user-modal-content, .pipeline-modal-content {
    margin: auto;
    padding: 10px;
  }

  .modal-footer {
    height: 45px;
    padding-top: 15px;
  }

</style>
