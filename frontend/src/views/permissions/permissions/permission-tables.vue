<template>
  <div>
    <div v-if="permissionOptions && permissionOptions.length > 0">
      <div v-for="category in permissionOptions" :key="category.name">
        <p style="margin-top: 20px;">{{ category.name }}: {{ category.description }}</p><br>
        <table class="table is-narrow is-fullwidth table-general">
          <thead>
          <tr>
            <th style="text-align: center" width="60"><input type="checkbox" @click="checkAll(category)"
                                                             :checked="allSelected(category)"/>
            </th>
            <th width="300">Name</th>
            <th>Description</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="role in category.roles">
            <td style="text-align: center"><input type="checkbox" :id="getFullName(category, role)"
                                                  :value="getFullName(category, role)"
                                                  v-model="roles"></td>
            <td>{{role.name}}</td>
            <td>{{role.description}}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'permission-tables',
  props: {
    value: Array,
    permissionOptions: Array
  },
  data () {
    return {
      roles: []
    }
  },
  watch: {
    roles: function (val) {
      this.$emit('input', val)
    },
    value: function (val) {
      this.roles = val === undefined ? [] : val
    }
  },
  methods: {
    checkAll (category) {
      if (this.allSelected(category)) {
        this.deselectAll(category)
      } else {
        this.selectAll(category)
      }
    },
    selectAll (category) {
      this.flattenOptions(category).forEach(p => {
        if (this.roles.indexOf(p) === -1) {
          this.roles.push(p)
        }
      })
    },
    deselectAll (category) {
      this.flattenOptions(category).forEach(p => {
        let index = this.roles.indexOf(p)
        if (index > -1) {
          this.roles.splice(index, 1)
        }
      })
    },
    allSelected (category) {
      for (let role of category.roles) {
        const name = this.getFullName(category, role)
        if (this.roles.indexOf(name) === -1) {
          return false
        }
      }
      return true
    },
    flattenOptions (category) {
      return category.roles.map(p => category.name + p.name)
    },
    getFullName (category, role) {
      return category.name + role.name
    }
  }
}
</script>

<style scoped>
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

  .table-users td:hover {
    border: 2px solid #000;
    background: #575463;
    cursor: pointer;
  }
</style>
