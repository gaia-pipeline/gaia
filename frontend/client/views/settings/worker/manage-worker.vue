<template>
  <div class="tile is-vertical is-ancestor">
    <div class="tile is-parent">
      <a class="button is-primary" style="margin-bottom: -10px;">
        <span class="icon">
          <i class="fa fa-trash"></i>
        </span>
        <span>Reset registration code</span>
      </a>
    </div>
    <div class="tile">
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Worker registration code: </span>
          <message :direction="'down'" :message="registerCode" :duration="0"></message>
        </article>
      </div>
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Number of active worker: </span><b>{{ statusView.activeworker }}</b><br />
          <span>Number of suspended worker: </span><b>{{ statusView.suspendedworker }}</b><br />
          <span>Number of inactive worker: </span><b>{{ statusView.inactiveworker }}</b>
        </article>
      </div>
      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
          <span>Finished pipeline runs by worker: </span><b>{{ statusView.finishedruns }}</b><br />
          <span>Pipeline queue size: </span><b>{{ statusView.queuesize }}</b>
        </article>
      </div>
    </div>
    <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <vue-good-table
          :columns="workerColumns"
          :rows="workerRows"
          :paginate="true"
          :global-search="true"
          :defaultSortBy="{field: 'name', type: 'desc'}"
          globalSearchPlaceholder="Search ..."
          styleClass="table table-grid table-own-bordered">
          <template slot="table-row" slot-scope="props">
            <td>
              <span>{{ props.row.name }}</span>
            </td>
            <td>
              <div v-if="props.row.status === 'active'" style="color: green;">{{ props.row.status }}</div>
              <div v-else-if="props.row.status === 'inactive'" style="color: red;">{{ props.row.status }}</div>
              <div v-else style="color: #4da2fc;">{{ props.row.status }}</div>
            </td>
            <td>
              {{ convertTime(props.row.registerdate) }}
            </td>
            <td>
              {{ convertTime(props.row.lastcontact) }}
            </td>
            <td>
              <a v-on:click="deregisterWorker(props.row.uniqueid)"><i class="fa fa-ban" style="color: whitesmoke;"></i></a>
            </td>
          </template>
          <div slot="emptystate" class="empty-table-text">
            No worker found.
          </div>
        </vue-good-table>
      </article>
    </div>
  </div>
</template>

<script>
  import Vue from 'vue'
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import VueGoodTable from 'vue-good-table'
  import moment from 'moment'
  import Message from 'vue-bulma-message-html'

  Vue.use(VueGoodTable)

  export default {
    name: 'manage-worker',
    components: {Tabs, TabPane, Message},
    data () {
      return {
        registerCode: '',
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
          }
        ],
        workerRows: []
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
          .get('/api/v1/worker/secret')
          .then(response => {
            if (response.data) {
              this.registerCode = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })

        // Get status overview of all workers
        this.$http
          .get('/api/v1/worker/status')
          .then(response => {
            if (response.data) {
              this.statusView = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })

        // Get worker
        this.$http
          .get('/api/v1/worker')
          .then(response => {
            if (response.data) {
              this.workerRows = response.data
            }
          })
          .catch((error) => {
            this.$onError(error)
          })
      },

      deregisterWorker (id) {
        this.$http
          .delete('/api/v1/worker/' + id)
          .catch((error) => {
            this.$onError(error)
          })
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
</style>
