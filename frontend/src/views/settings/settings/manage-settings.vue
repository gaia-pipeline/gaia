<template>
    <div class="tile is-vertical">
      <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <div class="tile is-parent">
          <article class="tile is-child notification content-article box">
            <vue-good-table
              :columns="settingColumns"
              :rows="settingRows"
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
                <span v-if="props.column.field === 'display_name'">
                  <span>{{ props.row.display_name }}</span>
                </span>
                <span v-if="props.column.field === 'display_value'" v-tippy="{ arrow : true,  animation : 'shift-away'}">
                  <toggle-button
                    v-model="props.row.display_value"
                    id="pollertoggle"
                    :color="{checked: '#7DCE94', unchecked: '#82C7EB'}"
                    :labels="{checked: 'On', unchecked: 'Off'}"
                    @change="settingsTogglePollerSwitch"
                    :sync="true"/>
                </span>
              </template>
              <div slot="emptystate" class="empty-table-text">
                No settings found.
              </div>
            </vue-good-table>
          </article>
        </div>
      </article>
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
import { ToggleButton } from 'vue-js-toggle-button'
import { VueGoodTable } from 'vue-good-table'
import 'vue-good-table/dist/vue-good-table.css'
import VueTippy from 'vue-tippy'
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

Vue.use(VueTippy)

export default {
  name: 'manage-settings',
  components: { ToggleButton, VueGoodTable },
  data () {
    return {
      settingsTogglePollerValue: false,
      settingColumns: [
        {
          label: 'Name',
          field: 'display_name'
        },
        {
          label: 'Value',
          field: 'display_value'
        }
      ],
      settingRows: []
    }
  },
  mounted () {
    this.setSettings()
  },
  methods: {
    settingsTogglePollerSwitch (val) {
      if (val.value) {
        this.$http
          .post('/api/v1/settings/poll/on')
          .then(response => {
            openNotification({
              title: 'Poll turned on!',
              message: 'Polling has been enabled.',
              type: 'success'
            })
          })
          .catch((error) => {
            this.$onError(error)
          })
      } else {
        this.$http
          .post('/api/v1/settings/poll/off')
          .then(response => {
            openNotification({
              title: 'Poll turned off!',
              message: 'Polling has been disabled.',
              type: 'success'
            })
          })
          .catch((error) => {
            this.$onError(error)
          })
      }
    },
    setSettings () {
      this.$http
        .get('/api/v1/settings/poll', { params: { hideProgressBar: true } })
        .then(response => {
          this.settingRows = [{
            display_name: 'Polling',
            display_value: response.data.Status
          }]
        })
        .catch((error) => {
          this.$onError(error)
        })
    }
  }
}
</script>

<style scoped>
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
</style>
