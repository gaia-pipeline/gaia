<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <tabs type="boxed" :is-fullwidth="true" alignment="centered" size="large">
        <tab-pane label="Manage Secrets" icon="fa fa-user-circle">
          <div class="tile is-ancestor">
            <div class="tile is-vertical">
              <div class="tile is-parent">
                <a class="button is-primary" v-on:click="addSecretModal" style="margin-bottom: -10px;">
                  <span class="icon">
                    <i class="fa fa-user-plus"></i>
                  </span>
                  <span>Add Secrets</span>
                </a>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification content-article box">
                  <vue-good-table
                    :columns="keyColumns"
                    :rows="keyRows"
                    :paginate="true"
                    :global-search="true"
                    :defaultSortBy="{field: 'key', type: 'desc'}"
                    globalSearchPlaceholder="Search ..."
                    styleClass="table table-own-bordered">
                    <template slot="table-row" slot-scope="props">
                      <td>
                        <span>{{ props.row.key }}</span>
                      </td>
                      <td v-tippy="{ arrow : true,  animation : 'shift-away'}">
                        <span>{{ props.row.value }}</span>
                      </td>
                      <td>
                        <a v-on:click="deleteSecretModal(props.row)"><i class="fa fa-trash" style="color: whitesmoke;"></i></a>
                      </td>
                    </template>
                    <div slot="emptystate" class="empty-table-text">
                      No secrets found.
                    </div>
                  </vue-good-table>
                </article>
              </div>
            </div>
          </div>
        </tab-pane>
      </tabs>
    </div>

    <!-- delete secret modal -->
    <modal :visible="showDeleteSecretModal" class="modal-z-index" @close="close">
      <div class="box secret-modal">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <p>
                <span style="color: whitesmoke;">Do you really want to delete the secret {{ selectSecret.key }}?</span>
              </p>
            </div>
            <div class="modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="deleteSecret" style="width:150px;">Yes</button>
              </div>
              <div style="float: right;">
                <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </modal>

    <!-- add secret modal -->
    <modal :visible="showAddSecretModal" class="modal-z-index" @close="close">
      <div class="box secret-modal">
        <div class="block secret-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Add Secret" selected>
              <div class="secret-modal-content">
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="text" v-model="selectSecret.key" placeholder="Key">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectSecret.value" placeholder="Secret">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectSecret.valueconf" placeholder="Secret confirmation">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="addSecret">Add Secret</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
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
import { mapGetters } from 'vuex'

const NotificationComponent = Vue.extend(Notification)
const openNotification = (
  propsData = {
    title: '',
    message: '',
    type: '',
    direction: '',
    duration: 4500,
    container: '.notifications'
  }
) => {
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
      keyColumns: [
        {
          label: 'Name',
          field: 'key'
        },
        {
          label: 'Value',
          field: 'secret_value'
        },
        {
          label: ''
        }
      ],
      keyRows: [],
      selectSecret: {},
      showDeleteSecretModal: false,
      showAddSecretModal: false
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
        .get('/api/v1/secrets', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.keyRows = response.data
          }
        })
        .catch((error) => {
          this.$onError(error)
        })
    },

    convertTime (time) {
      return moment(time).fromNow()
    },

    deleteSecretModal (secret) {
      this.selectSecret = secret
      this.showDeleteSecretModal = true
    },

    addSecretModal () {
      this.selectSecret = {}
      this.showAddSecretModal = true
    },

    close () {
      this.showDeleteSecretModal = false
      this.showAddSecretModal = false
      this.selectSecret = {}
      this.$emit('close')
    },

    addSecret () {
      // pre-validate
      if (!this.selectSecret.value || !this.selectSecret.valueconf) {
        openNotification({
          title: 'Empty value',
          message: 'Empty value is not allowed.',
          type: 'danger'
        })
        this.close()
        return
      }

      // pre-validate
      if (!this.selectSecret.key || this.selectSecret.key.trim() === '') {
        openNotification({
          title: 'Empty secret',
          message: 'Empty secret is not allowed.',
          type: 'danger'
        })
        this.close()
        return
      }

      // pre-validate
      if (this.selectSecret.value !== this.selectSecret.valueconf) {
        openNotification({
          title: 'value not identical',
          message: 'value and confirmation are not identical!',
          type: 'danger'
        })
        this.close()
        return
      }
      this.selectSecret.valueconf = null

      this.$http
        .post('/api/v1/secret', this.selectSecret)
        .then(response => {
          openNotification({
            title: 'Secret added!',
            message: 'Secret has been successfully added.',
            type: 'success'
          })
          this.fetchData()
        })
        .catch(error => {
          this.$onError(error)
        })
      this.close()
    },

    deleteSecret () {
      this.$http
        .delete('/api/v1/secret/' + this.selectSecret.key)
        .then(response => {
          openNotification({
            title: 'Secret deleted!',
            message:
              'Secret ' +
              this.selectSecret.key +
              ' has been successfully deleted.',
            type: 'success'
          })
          this.fetchData()
          this.close()
        })
        .catch(error => {
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

  .secret-modal {
    text-align: center;
    background-color: #2a2735;
  }

  .secret-modal-content {
    margin: auto;
    padding: 10px;
  }

  .modal-footer {
    height: 45px;
    padding-top: 15px;
  }
</style>
