<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile is-ancestor">
        <div class="tile is-vertical">
          <div class="tile is-parent">
            <a class="button is-primary" v-on:click="addSecretModal" style="margin-bottom: -10px;margin-top: 10px;">
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
                :pagination-options="{
                  enabled: true,
                  mode: 'records'
                }"
                :search-options="{enabled: true, placeholder: 'Search ...'}"
                :sort-options="{
                  enabled: true,
                  initialSortBy: {field: 'key', type: 'desc'}
                }"
                styleClass="table table-grid table-own-bordered">
                <template slot="table-row" slot-scope="props">
                  <span v-if="props.column.field === 'key'">
                    <span>{{ props.row.key }}</span>
                  </span>
                  <span v-if="props.column.field === 'secret_value'" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                    <span>*****</span>
                  </span>
                  <span v-if="props.column.field === 'action'">
                    <a v-on:click="editSecretModal(props.row)"><i class="fa fa-edit" style="color: whitesmoke;"></i></a>
                    <a v-on:click="deleteSecretModal(props.row)"><i class="fa fa-trash" style="color: whitesmoke;"></i></a>
                  </span>
                </template>
                <div slot="emptystate" class="empty-table-text">
                  No secrets found.
                </div>
              </vue-good-table>
            </article>
          </div>
        </div>
      </div>
    </div>

    <!-- edit secret modal -->
    <modal :visible="showEditSecretModal" class="modal-z-index" @close="close">
      <div class="box secret-modal">
        <div class="block secret-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Change Secret" selected>
              <div class="secret-modal-content">
                <label class="label" style="text-align: left;">Change secret value for key {{ selectSecret.key }}:</label>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectSecret.newvalue" placeholder="New Value">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="selectSecret.newvalueconf" placeholder="New Value confirmation">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="changeSecret">Change Secret Value</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

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
import { Modal } from 'vue-bulma-modal'
import { Collapse, Item as CollapseItem } from 'vue-bulma-collapse-fixed'
import { VueGoodTable } from 'vue-good-table'
import 'vue-good-table/dist/vue-good-table.css'
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

Vue.use(VueTippy)

export default {
  components: {
    Modal,
    Collapse,
    CollapseItem,
    VueGoodTable
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
          label: 'Action',
          field: 'action'
        }
      ],
      keyRows: [],
      selectSecret: {},
      showEditSecretModal: false,
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
        .get('/api/v1/secrets', { params: { hideProgressBar: true } })
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

    editSecretModal (secret) {
      this.selectSecret = secret
      this.showEditSecretModal = true
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
      this.showEditSecretModal = false
      this.showDeleteSecretModal = false
      this.showAddSecretModal = false
      this.selectSecret = {}
      this.$emit('close')
    },

    preValidate (value, valueconf) {
      if (!value || !valueconf) {
        openNotification({
          title: 'Empty value',
          message: 'Empty value is not allowed.',
          type: 'danger'
        })
        return false
      }

      // pre-validate
      if (value !== valueconf) {
        openNotification({
          title: 'value not identical',
          message: 'value and confirmation are not identical!',
          type: 'danger'
        })
        return false
      }
      return true
    },

    changeSecret () {
      if (!this.preValidate(this.selectSecret.newvalue, this.selectSecret.newvalueconf)) {
        this.close()
        return
      }

      this.selectSecret.value = null
      this.$http
        .put('/api/v1/secret/update', this.selectSecret)
        .then(response => {
          openNotification({
            title: 'Secret changed!',
            message: 'Secret has been successful changed.',
            type: 'success'
          })
        })
        .catch(error => {
          this.$onError(error)
        })
      this.close()
    },

    addSecret () {
      if (!this.preValidate(this.selectSecret.value, this.selectSecret.valueconf)) {
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
