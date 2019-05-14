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
              <span>{{ props.row.display_name }}</span>
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
            field: 'display_name'
          },
          {
            label: 'Value',
            field: 'display_value'
          }
        ],
        workerRows: []
      }
    },
    mounted () {
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
    },
    methods: {}
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
