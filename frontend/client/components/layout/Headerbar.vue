<template>
  <nav class="level app-levelbar">
    <div class="level-left">
      <div class="level-item">
        <h3 class="subtitle is-3">
          <strong>{{ name }}</strong>
        </h3>
      </div>
    </div>
  </nav>
</template>

<script>
export default {
  computed: {
    name () {
      return this.$route.name
    }
  }
}
</script>

<style lang="scss">

.subtitle strong {
  font-family: 'Lobster', 'Times', 'serif';
  font-size: 3rem;
  color: whitesmoke;
}
</style>
