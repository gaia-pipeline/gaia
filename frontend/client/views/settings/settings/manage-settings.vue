<template>
    <div class="tile is-vertical">
      <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <div class="tile is-parent">
          <article class="tile is-child notification content-article box">
            <table class="table table-grid table-own-bordered">
              <td>
                <p>
                  <font color="#eeeeee">Poller</font> <vb-switch id="pollertoggle" type="success" size="large" v-model="settingsTogglePollerValue" @change="settingsTogglePollerSwitch"/>
                </p>
              </td>
            </table>
          </article>
        </div>
      </article>
    </div>
  </div>
</template>

<script>
  import Vue from 'vue'
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import VbSwitch from 'vue-bulma-switch'
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
    components: {Tabs, TabPane, VbSwitch},
    data () {
      return {
        search: '',
        settingsTogglePollerValue: false,
        settingsTogglePollerText: 'Off'
      }
    },
    mounted () {
      this.setSettings()
    },
    methods: {
      settingsTogglePollerSwitch (val) {
        // TODO: Get and Send to API here.
        if (val) {
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
          this.close()
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
          this.close()
        }
      },
      setSettings () {
        this.$http
          .get('/api/v1/settings/poll', {showProgressBar: false})
          .then(response => {
            let poller = document.getElementById('pollertoggle')
            if (response.data.Status === true) {
              poller.parentElement.classList.add('checked')
            } else {
              poller.parentElement.classList.delete('checked')
            }
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
    border: 2px solid #000;
    background: #2c2b32;
    color: #4da2fc;
  }
  .table-general td {
    border: 2px solid #000;
    color: #8c91a0;
  }
  .table-settings td:hover {
    border: 2px solid #000;
    background: #575463;
    cursor: pointer;
  }
</style>
