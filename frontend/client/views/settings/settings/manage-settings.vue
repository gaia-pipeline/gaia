<template>
    <div class="tile is-vertical">
      <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <div class="tile is-parent">
          <article class="tile is-child notification content-article box">
            <table class="table-general responsive">
              <tr>
                <th>Setting</th>
                <th>Value</th>
              </tr>
              <tr>
                <td>
                  Poller
                </td>
                <td>
                  <toggle-button
                      v-model="settingsTogglePollerValue"
                      id="pollertoggle"
                      :color="{checked: '#7DCE94', unchecked: '#82C7EB'}"
                      :labels="{checked: 'On', unchecked: 'Off'}"
                      @change="settingsTogglePollerSwitch"
                      :sync="true"/>
                </td>
              </tr>
            </table>
          </article>
        </div>
      </article>
    </div>
  </div>
</template>

<script>
  import Vue from 'vue'
  import { ToggleButton } from 'vue-js-toggle-button'
  import {TabPane, Tabs} from 'vue-bulma-tabs'
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

  export default {
    name: 'manage-settings',
    components: {Tabs, TabPane, ToggleButton},
    data () {
      return {
        // search: '',
        settingsTogglePollerValue: false
      }
    },
    mounted () {
      this.setSettings()
    },
    methods: {
      settingsTogglePollerSwitch (val) {
        // TODO: Get and Send to API here.
        if (val.value) {
          // this.settingsTogglePollerText = 'On'
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
          .get('/api/v1/settings/poll', {showProgressBar: false})
          .then(response => {
            this.settingsTogglePollerValue = response.data.Status
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
