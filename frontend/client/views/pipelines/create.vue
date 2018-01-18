<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical is-5">
      <div class="tile">
        <div class="tile is-parent is-vertical">
          <article class="tile is-child notification content-article">
            <!--<p class="title title-text">Compile from source</p>-->
            <div class="content">
              <label class="label">Copy the link of your <strong>git repo</strong> here.</label>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" v-focus v-model.lazy="gitURL" type="text" placeholder="Link to git repo ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-git"></i>
                </span>
                <span style="color: red;" v-if="gitNeedAuth">You are not authorized. Invalid username and/or password!</span>
                <span style="color: red;" v-if="gitInvalidURL">Invalid link to git repo provided!</span>
                <div>
                  <span>Branch:</span>
                  <div class="select is-fullwidth" v-if="gitBranches.length > 0">
                    <select v-model="gitBranchSelected">
                      <option v-for="branch in gitBranches" :key="branch" :value="branch">{{ branch }}</option>
                    </select>
                  </div>
                </div>
              </p>
              <p class="control">
                <a class="button" v-on:click="showCredentialsModal">
                  <span class="icon">
                    <i class="fa fa-certificate"></i>
                  </span>
                  <span>Add credentials</span>
                </a>
              </p>
              <hr>
              <label class="label">Type the name of your pipeline. You can put your pipelines into folders by defining a path. For example <strong>MyFolder/MyAwesomePipeline</strong>.</label>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" v-model="pipelineName" type="text" placeholder="Pipeline name ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-book"></i>
                </span>
              </p>
              <hr>
              <a class="button">
                <span class="icon">
                  <i class="fa fa-plus"></i>
                </span>
                <span>Create Pipeline</span>
              </a>
            </div>
          </article>
        </div>
      </div>
    </div>

    <!-- Credentials modal -->
    <modal :visible="gitCredentialsModal" class="modal-z-index" @close="close">
      <div class="box credentials-modal">
        <div class="block credentials-modal-content">
          <div class="credentials-modal-content">
            <p class="control has-icons-left" style="padding-bottom: 5px;">
              <input class="input is-medium input-bar" v-focus type="text" v-model="gitUsername" placeholder="Username">
              <span class="icon is-small is-left">
                <i class="fa fa-user-circle"></i>
              </span>
            </p>
            <p class="control has-icons-left">
              <input class="input is-medium input-bar" type="password" v-model="gitPassword" placeholder="Password">
              <span class="icon is-small is-left">
                <i class="fa fa-lock"></i>
              </span>
            </p>
          </div>
          <hr>
          <div class="block credentials-modal-content">
            <p class="control">
              <label class="label">SSH Private Key:</label>
              <textarea class="textarea input-bar" v-model="gitPrivateKey"></textarea>
            </p>
          </div>
          <div class="credentials-modal-content">
            <button class="button is-primary" v-on:click="close">Add Credentials</button>
          </div>
        </div>
      </div> 
    </modal>
  </div>
</template>

<script>
import { Modal } from 'vue-bulma-modal'

export default {

  data () {
    return {
      gitURL: '',
      gitNeedAuth: false,
      gitInvalidURL: false,
      gitCredentialsModal: false,
      gitUsername: '',
      gitPassword: '',
      gitPrivateKey: '',
      gitBranches: [
        'Master'
      ],
      gitBranchSelected: '',
      pipelineName: ''
    }
  },

  components: {
    Modal
  },

  watch: {
    gitURL: function () {
      // lets check if we can access the git repo
      var gitrepo = {
        giturl: this.gitURL,
        gituser: this.gitUsername,
        gitpassword: this.gitPassword
      }

      this.$http.post('/api/v1/pipelines/gitlsremote', gitrepo)
      .then((response) => {
        // Reset error message before
        this.gitNeedAuth = false
        this.gitInvalidURL = false

        // Get branches and set to master if available
        this.gitBranches = response.data.branches
        for (var i = 0; i < this.gitBranches.length; i++) {
          if (this.gitBranches[i] === 'refs/heads/master') {
            this.gitBranchSelected = this.gitBranches[i]
          }
        }

        // if we cannot find master
        if (!this.gitBranchSelected) {
          this.gitBranchSelected = this.gitBranches[0]
        }

        console.log(response.data)
      })
      .catch((error) => {
        // Need authentication
        if (error.response && error.response.status === 403) {
          this.gitNeedAuth = true
          this.gitInvalidURL = false
        } else if (error.response && error.response.status === 400) {
          this.gitInvalidURL = true
          this.gitNeedAuth = false
        }
        console.log(error.response.data)
      })
    }
  },

  methods: {
    close () {
      this.gitCredentialsModal = false
      this.$emit('close')
    },

    showCredentialsModal () {
      this.gitCredentialsModal = true
    }
  }

}
</script>

<style lang="scss" scoped>

.credentials-modal {
  text-align: center;
  background-color: #2a2735;
}

.credentials-modal-content {
  margin: auto;
  padding: 10px;
}

</style>
