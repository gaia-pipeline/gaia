<template>
    <div class="tile is-vertical">
      <div class="tile is-parent">
      <article class="tile is-child notification content-article box">
        <div class="tile is-parent">
          <article class="tile is-child notification content-article box">
            <table class="table table-grid table-own-bordered">
              <td>
                <p>
                  <font color="#eeeeee">Poller</font> <vb-switch type="success" size="large" v-model="settingsTogglePollerValue" @change="settingsTogglePollerSwitch"/>
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
  import {TabPane, Tabs} from 'vue-bulma-tabs'
  import VbSwitch from 'vue-bulma-switch'

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
          this.settingsTogglePollerText = 'On'
        } else {
          this.settingsTogglePollerText = 'Off'
        }
      },
      setSettings () {
        // TODO: Setup the toggles and settings to the current settings.
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
