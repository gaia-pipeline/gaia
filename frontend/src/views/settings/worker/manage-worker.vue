<template>
  <div class="tile is-vertical is-ancestor">
    <div class="tile is-parent">
      <a class="button is-primary" style="margin-bottom: -10px;" v-on:click="showResetSecretModal">
        <span class="icon">
          <i class="fa fa-trash"></i>
        </span>
        <span>Reset registration secret</span>
      </a>
    </div>
    <div class="tile">
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Worker registration secret: </span>
          <message :direction="'down'" :message="registerSecret" :duration="0"></message>
        </article>
      </div>
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Number of active worker: </span><b>{{ statusView.activeworker }}</b><br/>
          <span>Number of suspended worker: </span><b>{{ statusView.suspendedworker }}</b><br/>
          <span>Number of inactive worker: </span><b>{{ statusView.inactiveworker }}</b>
        </article>
      </div>
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Finished pipeline runs by worker: </span><b>{{ statusView.finishedruns }}</b><br/>
          <span>Pipeline queue size: </span><b>{{ statusView.queuesize }}</b>
        </article>
      </div>
    </div>
    <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <vue-good-table
          :columns="workerColumns"
          :rows="workerRows"
          :pagination-options="{
            enabled: true,
            mode: 'records'
          }"
          :search-options="{enabled: true, placeholder: 'Search ...'}"
          :sort-options="{
            enabled: true,
            initialSortBy: {field: 'name', type: 'desc'}
          }"
          styleClass="table table-grid table-own-bordered">
          <template slot="table-row" slot-scope="props">
            <span v-if="props.column.field === 'name'">
              <span>{{ props.row.name }}</span>
            </span>
            <span v-if="props.column.field === 'status'">
              <div v-if="props.row.status === 'active'" style="color: green;">{{ props.row.status }}</div>
              <div v-else-if="props.row.status === 'inactive'" style="color: red;">{{ props.row.status }}</div>
              <div v-else style="color: #4da2fc;">{{ props.row.status }}</div>
            </span>
            <span v-if="props.column.field === 'registerdate'">
              <span :title="props.row.registerdate" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                {{ convertTime(props.row.registerdate) }}
              </span>
            </span>
            <span v-if="props.column.field === 'lastcontact'">
              <span :title="props.row.lastcontact" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                {{ convertTime(props.row.lastcontact) }}
              </span>
            </span>
            <span v-if="props.column.field === 'tags'">
              <span>{{ $prettifyTags(props.row.tags.sort()) }}</span>
            </span>
            <span v-if="props.column.field === 'action'">
              <a title="Deregister Worker" v-tippy="{ arrow : true,  animation : 'shift-away'}"
                 v-on:click="deregisterWorkerModal(props.row)"><i class="fa fa-ban"
                                                                  style="color: whitesmoke;"></i></a>
            </span>
          </template>
          <div slot="emptystate" class="empty-table-text">
            No worker found.
          </div>
        </vue-good-table>
      </article>
    </div>

    <!-- deregister worker modal -->
    <modal :visible="showDeregisterWorkerModal" class="modal-z-index" @close="close">
      <div class="box confirmation-modal">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <p>
                <span
                  style="color: whitesmoke;">Do you really want to deregister the worker {{ selectedWorker.name }}?</span>
              </p>
            </div>
            <div class="confirmation-modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="deregisterWorker" style="width:150px;">Yes</button>
              </div>
              <div style="float: right;">
                <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
              </div>
            </div>
          </div>
        </article>
      </div>
    </modal>

    <!-- reset global worker registration secret modal -->
    <modal :visible="showResetWorkerSecretModal" class="modal-z-index" @close="close">
      <div class="box confirmation-modal">
        <article class="media">
          <div class="media-content">
            <div class="content">
              <p>
                <span style="color: whitesmoke;">
                  This will reset the global worker registration secret and generates a new one.
                  Workers which are already registered will not be affected by this action.
                  Do you really want to reset the global worker registration secret?
                </span>
              </p>
            </div>
            <div class="confirmation-modal-footer">
              <div style="float: left;">
                <button class="button is-primary" v-on:click="resetWorkerSecret" style="width:150px;">Yes</button>
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
import { VueGoodTable } from 'vue-good-table'
import 'vue-good-table/dist/vue-good-table.css'
import { Modal } from 'vue-bulma-modal'
import moment from 'moment'
import Message from 'vue-bulma-message-html'
import VueTippy from 'vue-tippy'
import Notification from 'vue-bulma-notification-fixed'

Vue.use(VueTippy)

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

export default {
  name: 'manage-worker',
  components: { Message, Modal, VueGoodTable },
  data () {
    return {
      registerSecret: '',
      statusView: {},
      workerColumns: [
        {
          label: 'Name',
          field: 'name'
        },
        {
          label: 'Status',
          field: 'status'
        },
        {
          label: 'Register date',
          field: 'registerdate'
        },
        {
          label: 'Last contact',
          field: 'lastcontact'
        },
        {
          label: 'Tags',
          field: 'tags'
        },
        {
          label: 'Action',
          field: 'action'
        }
      ],
      workerRows: [],
      showDeregisterWorkerModal: false,
      showResetWorkerSecretModal: false,
      selectedWorker: {}
    }
  },
  mounted () {
    // fetch data from API
    this.fetchData()

    // periodically fetch updated data
    let intervalID = setInterval(function () {
      this.fetchData()
    }.bind(this), 3000)

    // Append interval id to store
    this.$store.commit('appendInterval', intervalID)
  },
  destroyed () {
    this.$store.commit('clearIntervals')
  },
  watch: {
    '$route': 'fetchData'
  },
  methods: {
    fetchData () {
      // Get registration code for new worker
      this.$http
        .get('/api/v1/worker/secret', { params: { hideProgressBar: true } })
        .then(response => {
          if (response.data) {
            this.registerSecret = response.data
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })

      // Get status overview of all workers
      this.$http
        .get('/api/v1/worker/status', { params: { hideProgressBar: true } })
        .then(response => {
          if (response.data) {
            this.statusView = response.data
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })

      // Get worker
      this.$http
        .get('/api/v1/worker', { params: { hideProgressBar: true } })
        .then(response => {
          if (response.data) {
            this.workerRows = response.data
          } else {
            this.workerRows = []
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })
    },
    deregisterWorker () {
      this.$http
        .delete('/api/v1/worker/' + this.selectedWorker.uniqueid)
        .then(response => {
          openNotification({
            title: 'Worker deregistered!',
            message:
                'Worker ' +
                this.selectedWorker.name +
                ' has been successfully deregistered.',
            type: 'success'
          })
          this.fetchData()
          this.close()
        })
        .catch((error) => {
          this.$onError(error)
        })
    },
    resetWorkerSecret () {
      this.$http
        .post('/api/v1/worker/secret')
        .then(response => {
          openNotification({
            title: 'Secret reset successful',
            message: 'Successfully generated and stored a new global worker registration secret.',
            type: 'success'
          })
          this.fetchData()
          this.close()
        })
        .catch((error) => {
          this.$onError(error)
        })
    },
    deregisterWorkerModal (worker) {
      this.selectedWorker = worker
      this.showDeregisterWorkerModal = true
    },
    showResetSecretModal () {
      this.showResetWorkerSecretModal = true
    },
    close () {
      this.selectedWorker = {}
      this.showDeregisterWorkerModal = false
      this.showResetWorkerSecretModal = false
    },
    convertTime (time) {
      return moment(time).fromNow()
    }
  }
}
</script>

<style>
  .settings-row {
    cursor: pointer;
  }

  .table-general {
    background: #413F4A;
    border: 2px solid #000;
  }

  .table-general th {
    color: #4da2fc;
  }

  .table-general td {
    border: 2px solid #000;
    color: #8c91a0;
  }

  .table-settings td:hover {
    background: #575463;
    cursor: pointer;
  }

  .message-header {
    background-color: #4da2fc;
  }

  .message-body {
    background-color: black;
    border: none;
    color: whitesmoke;
  }

  .confirmation-modal {
    text-align: center;
    background-color: #2a2735;
  }

  .confirmation-modal-footer {
    height: 45px;
    padding-top: 15px;
  }
</style>
