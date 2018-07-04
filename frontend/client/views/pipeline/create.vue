<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile">
        <div class="tile is-vertical is-parent is-6">
          <article class="tile is-child notification content-article">
            <div class="content">
              <label class="label">Copy the link of your
                <strong>git repo</strong> here.</label>
              <p class="control has-icons-left" v-bind:class="{ 'has-icons-right': gitSuccess }">
                <input class="input is-medium input-bar" v-focus v-model="giturl" v-on:input="checkGitRepoDebounce" type="text" placeholder="Link to git repo ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-git"></i>
                </span>
                <span v-if="gitSuccess" class="icon is-small is-right is-blue">
                  <i class="fa fa-check"></i>
                </span>
              </p>
              <span style="color: red" v-if="gitErrorMsg">Cannot access git repo: {{ gitErrorMsg }}</span>
              <div v-if="gitBranches.length > 0">
                <span>Branch:</span>
                <div class="select is-fullwidth">
                  <select v-model="createPipeline.pipeline.repo.selectedbranch">
                    <option v-for="branch in gitBranches" :key="branch" :value="branch">{{ branch }}</option>
                  </select>
                </div>
              </div>
              <p class="control" style="padding-top: 10px;">
                <a class="button is-primary" v-on:click="showCredentialsModal">
                  <span class="icon">
                    <i class="fa fa-certificate"></i>
                  </span>
                  <span>Add credentials</span>
                </a>
              </p>
              <hr class="dotted-line">
              <label class="label">Type the name of your pipeline here.</label>
              <p class="control has-icons-left" v-bind:class="{ 'has-icons-right': pipelineNameSuccess }">
                <input class="input is-medium input-bar" v-model="pipelinename" v-on:input="checkPipelineNameAvailableDebounce" type="text" placeholder="Pipeline name ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-book"></i>
                </span>
                <span v-if="pipelineNameSuccess" class="icon is-small is-right is-blue">
                  <i class="fa fa-check"></i>
                </span>
              </p>
              <span style="color: red" v-if="pipelineErrorMsg">Pipeline Name incorrect: {{ pipelineErrorMsg }}</span>
              <hr class="dotted-line">
              <a class="button is-green-button" v-on:click="startCreatePipeline" v-bind:class="{ 'is-disabled': !gitSuccess || !pipelineNameSuccess }">
                <span class="icon">
                  <i class="fa fa-plus"></i>
                </span>
                <span>Create Pipeline</span>
              </a>
            </div>
          </article>
        </div>

        <div class="tile is-parent is-4">
          <article class="tile is-child notification content-article box">
            <p class="subtitle">Select pipeline language</p>
            <div class="content" style="display: flex;">
              <div class="pipelinetype" title="Golang" v-tippy="{ arrow : true,  animation : 'shift-away'}" v-on:click="createPipeline.pipeline.type = 'golang'" v-bind:class="{ pipelinetypeactive: createPipeline.pipeline.type === 'golang' }" data-tippy-hideOnClick="false">
                <img src="~assets/golang.png" class="typeimage">
              </div>
              <div class="pipelinetype" title="Python (not yet supported)" v-tippy="{ arrow : true,  animation : 'shift-away'}" v-bind:class="{ pipelinetypeactive: createPipeline.pipeline.type === 'python' }" data-tippy-hideOnClick="false">
                <img src="~assets/python.png" class="typeimage typeimagenotyetsupported">
              </div>
              <div class="pipelinetype" title="Java (not yet supported)" v-tippy="{ arrow : true,  animation : 'shift-away'}" v-bind:class="{ pipelinetypeactive: createPipeline.pipeline.type === 'java' }" data-tippy-hideOnClick="false">
                <img src="~assets/java.png" class="typeimage typeimagenotyetsupported">
              </div>
            </div>
            <div class="content" style="display: flex;">
              <div class="pipelinetype" title="C++ (not yet supported)" v-tippy="{ arrow : true,  animation : 'shift-away'}" v-bind:class="{ pipelinetypeactive: createPipeline.pipeline.type === 'cplusplus' }" data-tippy-hideOnClick="false">
                <img src="~assets/cplusplus.png" class="typeimage typeimagenotyetsupported">
              </div>
              <div class="pipelinetype" title="Node.js (not yet supported)" v-tippy="{ arrow : true,  animation : 'shift-away'}" v-bind:class="{ pipelinetypeactive: createPipeline.pipeline.type === 'nodejs' }" data-tippy-hideOnClick="false">
                <img src="~assets/nodejs.png" class="typeimage typeimagenotyetsupported">
              </div>
            </div>
          </article>
        </div>
      </div>

      <div class="tile is-parent is-10">
        <article class="tile is-child notification content-article box">
            <vue-good-table
              title="Pipeline history"
              :columns="historyColumns"
              :rows="historyRows"
              :paginate="true"
              :global-search="true"
              :defaultSortBy="{field: 'status', type: 'asc'}"
              globalSearchPlaceholder="Search ..."
              styleClass="table table-own-bordered">
              <template slot="table-row" slot-scope="props">
                <td>{{ props.row.pipeline.name }}</td>
                <td class="progress-bar-height">
                  <div class="progress-bar-middle blink" v-if="props.row.statustype === 'running'">
                    <progress-bar :type="'info'" :size="'small'" :value="props.row.status" :max="100" :show-label="false"></progress-bar>
                  </div>
                  <div v-else-if="props.row.statustype === 'success'" style="color: green;">{{ props.row.statustype }}</div>
                  <div v-else style="color: red;">{{ props.row.statustype }}</div>
                </td>
                <td>{{ props.row.pipeline.type }}</td>
                <td :title="props.row.created" v-tippy="{ arrow : true,  animation : 'shift-away'}">{{ convertTime(props.row.created) }}</td>
                <td>
                  <a class="button is-green-button is-small" @click="showStatusOutputModal(props.row.output)">
                    <span class="icon">
                      <i class="fa fa-align-justify"></i>
                    </span>
                    <span>Show Output</span>
                  </a>
                </td>
              </template>
              <div slot="emptystate" class="empty-table-text">
                No pipelines found in database.
              </div>
            </vue-good-table>
        </article>
      </div>
    </div>

    <!-- Credentials modal -->
    <modal :visible="gitCredentialsModal" class="modal-z-index" @close="close">
      <div class="box credentials-modal">
        <div class="block credentials-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Basic Authentication" selected>
              <div class="credentials-modal-content">
                <label class="label" style="text-align: left;">Add credentials for basic authentication:</label>
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="text" v-model="createPipeline.pipeline.repo.user" placeholder="Username">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user-circle"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="createPipeline.pipeline.repo.password" placeholder="Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
            <collapse-item title="SSH Key">
              <label class="label" style="text-align: left;">Instead of using basic authentication, provide a pem encoded private key.</label>
              <div class="block credentials-modal-content">
                <p class="control">
                  <textarea class="textarea input-bar" v-model="createPipeline.pipeline.repo.privatekey.key"></textarea>
                </p>
              </div>
              <h2 class="separater">
                <span class="span">Additional:</span>
              </h2>
              <p class="control has-icons-left" style="padding-bottom: 5px;">
                <input class="input is-medium input-bar" v-focus type="text" v-model="createPipeline.pipeline.repo.privatekey.username" placeholder="Username">
                <span class="icon is-small is-left">
                  <i class="fa fa-user-circle"></i>
                </span>
              </p>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" type="password" v-model="createPipeline.pipeline.repo.privatekey.password" placeholder="Password">
                <span class="icon is-small is-left">
                  <i class="fa fa-lock"></i>
                </span>
              </p>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="close">Add Credentials</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="cancel">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>

    <!-- status output modal -->
    <modal :visible="statusOutputModal" class="modal-z-index" @close="closeStatusModal">
      <div class="box statusModal">
        <div>
          <message :direction="'down'" :message="statusOutputMsg" :duration="0"></message>
        </div>
        <div class="modal-footer">
          <div style="float: right;">
            <button class="button is-danger" @click="closeStatusModal">Close</button>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<script>
import Vue from 'vue'
import { Modal } from 'vue-bulma-modal'
import { Collapse, Item as CollapseItem } from 'vue-bulma-collapse'
import ProgressBar from 'vue-bulma-progress-bar'
import VueTippy from 'vue-tippy'
import VueGoodTable from 'vue-good-table'
import moment from 'moment'
import Message from 'vue-bulma-message-html'

Vue.use(VueGoodTable)
Vue.use(VueTippy)

export default {
  data () {
    return {
      gitErrorMsg: '',
      gitSuccess: false,
      gitCredentialsModal: false,
      gitBranches: [],
      giturl: '',
      pipelinename: '',
      pipelineNameSuccess: false,
      pipelineErrorMsg: '',
      createPipeline: {
        id: '',
        output: '',
        status: 0,
        created: new Date(),
        pipeline: {
          name: '',
          type: 'golang',
          repo: {
            url: '',
            user: '',
            password: '',
            selectedbranch: '',
            privatekey: {
              key: '',
              username: '',
              password: ''
            }
          }
        }
      },
      historyColumns: [
        {
          label: 'Name',
          field: 'pipeline.name'
        },
        {
          label: 'Status',
          field: 'status',
          type: 'number'
        },
        {
          label: 'Type',
          field: 'pipeline.type'
        },
        {
          label: 'Creation date',
          field: 'created'
        },
        {
          label: 'Status Output',
          field: 'output'
        }
      ],
      historyRows: [],
      statusOutputModal: false,
      statusOutputMsg: ''
    }
  },

  components: {
    Modal,
    Collapse,
    CollapseItem,
    ProgressBar,
    Message
  },

  mounted () {
    // created pipelines history
    this.fetchData()

    // periodically update history dashboard
    var intervalID = setInterval(function () {
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
      this.$http
        .get('/api/v1/pipeline/created', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.historyRows = response.data

            // Add extra output messages if the pipeline is in creation status.
            for (var i = 0; i < this.historyRows.length; i++) {
              if (this.historyRows[i].statustype === 'running' || this.historyRows[i].statustype === 'success') {
                if (this.historyRows[i].status >= 25) {
                  this.historyRows[i].output = 'Git repository has been cloned.\n'
                  this.historyRows[i].output += 'Starting build process. This may take some time...\n'
                }

                if (this.historyRows[i].status >= 75) {
                  this.historyRows[i].output += 'Pipeline has been successfully compiled.\n'
                  this.historyRows[i].output += 'Copy binary to pipelines folder...\n'
                }

                if (this.historyRows[i].status === 100) {
                  this.historyRows[i].output += 'Copy was successful.\n'
                  this.historyRows[i].output += 'Finished.\n'
                }
              }
            }
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })
    },

    checkGitRepoDebounce: Vue._.debounce(function () {
      this.checkGitRepo()
    }, 500),

    checkGitRepo () {
      if (this.giturl === '') {
        return
      }

      // copy giturl into our struct
      this.createPipeline.pipeline.repo.url = this.giturl

      // Reset last fetches
      this.gitBranches = []
      this.createPipeline.pipeline.repo.selectedbranch = ''
      this.gitSuccess = false

      this.$http
        .post('/api/v1/pipeline/gitlsremote', this.createPipeline.pipeline.repo)
        .then(response => {
          // Reset error message before
          this.gitErrorMsg = ''
          this.gitSuccess = true

          // Get branches and set to master if available
          this.gitBranches = response.data
          for (var i = 0; i < this.gitBranches.length; i++) {
            if (this.gitBranches[i] === 'refs/heads/master') {
              this.createPipeline.pipeline.repo.selectedbranch = this.gitBranches[i]
            }
          }

          // if we cannot find master
          if (
            !this.createPipeline.pipeline.repo.selectedbranch &&
            this.gitBranches.length > 0
          ) {
            this.createPipeline.pipeline.repo.selectedbranch = this.gitBranches[0]
          }
        })
        .catch(error => {
          // Add error message
          this.gitErrorMsg = error.response.data
        })
    },

    checkPipelineNameAvailableDebounce: Vue._.debounce(function () {
      this.checkPipelineNameAvailable()
    }, 500),

    checkPipelineNameAvailable () {
      // copy pipeline name into struct
      this.createPipeline.pipeline.name = this.pipelinename

      // Request for availability
      this.$http
        .get('/api/v1/pipeline/name', { params: { name: this.pipelinename } })
        .then(response => {
          // pipeline name valid and available
          this.pipelineErrorMsg = ''
          this.pipelineNameSuccess = true
        })
        .catch(error => {
          this.pipelineErrorMsg = error.response.data
          this.pipelineNameSuccess = false
        })
    },

    startCreatePipeline () {
      // copy giturl into our struct and pipeline name
      this.createPipeline.pipeline.repo.url = this.giturl
      this.createPipeline.pipeline.name = this.pipelinename

      // Start the create pipeline process in the backend
      this.$http
        .post('/api/v1/pipeline', this.createPipeline)
        .then(response => {
          // Run fetchData to see the pipeline in our history table
          this.fetchData()
        })
        .catch((error) => {
          this.$onError(error)
        })
    },

    convertTime (time) {
      return moment(time).fromNow()
    },

    close () {
      this.checkGitRepo()
      this.gitCredentialsModal = false
      this.$emit('close')
    },

    cancel () {
      // cancel means reset all stuff
      this.createPipeline.pipeline.repo.user = ''
      this.createPipeline.pipeline.repo.password = ''
      this.createPipeline.pipeline.repo.privatekey.key = ''
      this.createPipeline.pipeline.repo.privatekey.username = ''
      this.createPipeline.pipeline.repo.privatekey.password = ''

      this.close()
    },

    showCredentialsModal () {
      this.gitCredentialsModal = true
    },

    showStatusOutputModal (msg) {
      if (!msg) {
        msg = 'No output found.'
      }

      // LF does not work for HTML. Replace with <br />
      msg = msg.replace(/\n/g, '<br />')

      this.statusOutputMsg = msg
      this.statusOutputModal = true
    },

    closeStatusModal () {
      this.statusOutputModal = false
      this.statusOutputMsg = ''
      this.$emit('close')
    }
  }
}
</script>

<style lang="scss">
.credentials-modal {
  text-align: center;
  background-color: #2a2735;
}

.credentials-modal-content {
  margin: auto;
  padding: 10px;
}

.dotted-line {
  background-image: linear-gradient(
    to right,
    black 33%,
    rgba(255, 255, 255, 0) 0%
  );
  background-position: bottom;
  background-size: 3px 1px;
  background-repeat: repeat-x;
}

.separater {
  width: 100%;
  text-align: center;
  border-bottom: 1px solid #4da2fc;
  line-height: 0.1em;
  padding-top: 15px;
  margin: 10px 0 20px;
}

.separater .span {
  background: black;
  color: whitesmoke;
  padding: 0 10px;
}

.modal-footer {
  height: 35px;
  padding-top: 15px;
}

.pipelinetype {
  height: 80px;
  width: 80px;
  border: 1px solid;
  color: black;
  margin: 0 5px;
  box-shadow: 4px 4px 4px 4px rgba(0, 0, 0, 0.2),
    0 1px 1px 0 rgba(0, 0, 0, 0.19);
}

.pipelinetype:hover {
  -moz-transform: scale(1.1);
  -webkit-transform: scale(1.1);
  transform: scale(1.1);
}

.pipelinetypeactive {
  color: #4da2fc;
  border: 2px solid;
}

.typeimage {
  display: block;
  margin: 0 auto;
  height: 100%;
}

.typeimagenotyetsupported {
  opacity: 0.5;
}

.blink {
  animation: blink 700ms infinite alternate;
}

@keyframes blink {
  from {
    opacity: 1;
  }
  to {
    opacity: 0.2;
  }
}

.message-header {
  background-color: #4da2fc;
}

.message-body {
  background-color: black;
  border: none;
  color: whitesmoke;
}

.statusModal {
  width: 100%;
  background-color: rgb(42, 38, 53);
}
</style>
