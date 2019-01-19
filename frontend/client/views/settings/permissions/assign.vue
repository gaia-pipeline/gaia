<template>
  <modal :visible="visible" class="modal-z-index" @close="close">
    <div class="box pipeline-modal">
      <article class="media">
        <div class="media-content">
          <div class="content">
            <p>
              Test...
            </p>
          </div>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="deletePipeline" style="width:150px;">Yes</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </modal>
</template>

<script>
  export default {
    name: 'assign',
    props: {
      visible: false
    }
  }
</script>

<style scoped>

</style>
