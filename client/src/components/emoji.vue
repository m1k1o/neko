<template>
  <div class="neko-emoji" v-on-clickaway="onClickAway">
    <div class="search">
      <div class="search-contianer">
        <input type="text" ref="search" v-model="search" />
      </div>
    </div>
    <div class="list" ref="scroll" @scroll="onScroll">
      <ul :class="['group-list']" :style="{ display: search === '' ? 'flex' : 'none' }">
        <li v-for="(group, index) in groups" :key="index" class="group" ref="groups">
          <span class="label">{{ group.name }}</span>
          <ul class="emoji-list">
            <li
              v-for="emoji in index === 0 ? recent : group.list"
              :key="`${group.id}-${emoji}`"
              :class="['emoji-container', hovered === emoji ? 'active' : '']"
            >
              <span
                :class="['emoji']"
                @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
                @click.stop.prevent="onClick($event, emoji)"
                :data-emoji="emoji"
              ></span>
            </li>
          </ul>
        </li>
      </ul>
      <ul :class="['emoji-container']" :style="{ display: search === '' ? 'none' : 'flex' }">
        <li v-for="emoji in filtered" :key="emoji" :class="['emoji-item', hovered === emoji ? 'active' : '']">
          <span
            :class="['emoji']"
            @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
            @click.stop.prevent="onClick($event, emoji)"
            :data-emoji="emoji"
          ></span>
        </li>
      </ul>
    </div>
    <div class="details">
      <div class="details-container" v-if="hovered !== ''">
        <span :class="['emoji']" :data-emoji="hovered" /><span class="emoji-id">:{{ hovered }}:</span>
      </div>
    </div>
    <div class="groups">
      <ul>
        <li
          v-for="(group, index) in groups"
          :key="index"
          :class="[group.id, active.id === group.id && search === '' ? 'active' : '']"
          @click.stop.prevent="scrollTo($event, index)"
        >
          <span :class="[`group-${group.id} fas`]" />
        </li>
      </ul>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  $emoji-width: 300px;

  .neko-emoji {
    position: absolute;
    z-index: 10000;
    width: $emoji-width;
    height: 350px;
    background: $background-secondary;
    bottom: 75px;
    right: 5px;
    display: flex;
    flex-direction: column;
    border-radius: 5px;
    overflow: hidden;
    box-shadow: $elevation-high;

    .search {
      flex-shrink: 0;
      border-bottom: 1px solid $background-tertiary;
      padding: 10px;

      .search-contianer {
        border-radius: 5px;
        color: $interactive-normal;
        position: relative;
        display: flex;
        flex-direction: column;
        align-content: center;
        overflow: hidden;

        &::before {
          content: '\f002';
          font-weight: 900;
          font-family: 'Font Awesome 6 Free';
          position: absolute;
          width: 15px;
          height: 15px;
          top: 6px;
          right: 6px;
          opacity: 0.5;
        }

        input {
          border: none;
          background-color: $background-floating;
          color: $interactive-normal;
          padding: 5px;
          font-weight: 500;

          &::placeholder {
            color: $text-muted;
            font-weight: 500;
          }
        }
      }
    }

    .list {
      position: relative;
      flex-grow: 1;
      overflow-y: scroll;
      overflow-x: hidden;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;
      scroll-behavior: smooth;
      padding: 5px;

      &::-webkit-scrollbar {
        width: 4px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border-radius: 4px;
      }

      &::-webkit-scrollbar-thumb:hover {
        background-color: $background-floating;
      }

      .group-list {
        width: $emoji-width;
        display: flex;
        flex-direction: column;

        li {
          &.group {
            .label {
              z-index: 2;
              text-transform: uppercase;
              font-weight: 500;
              font-size: 12px;
              position: sticky;
              top: -5px;
              background-color: rgba($color: $background-secondary, $alpha: 0.9);
              width: 100%;
              display: block;
              padding: 8px 0;
            }
          }
        }
      }

      .emoji-list {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        li {
          &.emoji-container {
            padding: 2px;
            border-radius: 3px;
            cursor: pointer;

            &.active {
              background-color: $background-floating;
            }
          }
        }
      }
    }

    .details {
      flex-shrink: 0;
      display: flex;
      align-content: center;
      justify-content: center;
      flex-direction: column;
      height: 36px;
      background: $background-tertiary;

      .details-container {
        display: flex;
        align-content: center;
        flex-direction: row;
        height: 20px;

        span {
          cursor: default;

          &.emoji {
            margin: 0 5px 0 10px;
          }

          &.emoji-id {
            line-height: 20px;
            font-size: 16px;
            font-weight: 500;
          }
        }
      }
    }

    .groups {
      flex-shrink: 0;
      height: 30px;
      background: $background-floating;
      padding: 0 5px;

      ul {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;

        li {
          flex-grow: 1;
          display: flex;
          flex-direction: row;
          justify-content: center;
          align-content: center;
          flex-direction: column;
          height: 27px;
          cursor: pointer;

          &.active {
            border-bottom: 3px solid $style-primary;
          }

          span {
            margin: 0 auto;
            height: 20px;
            width: 20px;
            font-size: 16px;
            line-height: 20px;
            text-align: center;

            &.group-recent::before {
              content: '\f017';
            }
            &.group-neko::before {
              content: '\f6be';
            }
            &.group-emotion::before {
              content: '\f118';
            }
            &.group-people::before {
              content: '\f0c0';
            }
            &.group-nature::before {
              content: '\f1b0';
            }
            &.group-food::before {
              content: '\f5d1';
            }
            &.group-activity::before {
              content: '\f44e';
            }
            &.group-travel::before {
              content: '\f1b9';
            }
            &.group-objects::before {
              content: '\f0eb';
            }
            &.group-symbols::before {
              content: '\f86d';
            }
            &.group-flags::before {
              content: '\f024';
            }
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'
  import { directive as onClickaway } from 'vue-clickaway'
  import { get } from '../utils/localstorage'

  @Component({
    name: 'neko-emoji',
    directives: {
      onClickaway,
    },
  })
  export default class extends Vue {
    @Ref('scroll') readonly _scroll!: HTMLElement
    @Ref('search') readonly _search!: HTMLInputElement
    @Ref('groups') readonly _groups!: HTMLElement[]

    waitingForPaint = false
    search = ''
    index = 0
    hovered = ''
    recent: string[] = JSON.parse(get('emoji_recent', '[]'))

    get active() {
      return this.$accessor.emoji.groups[this.index]
    }

    get keywords() {
      return this.$accessor.emoji.keywords
    }

    get groups() {
      return this.$accessor.emoji.groups
    }

    get list() {
      return this.$accessor.emoji.list
    }

    get filtered() {
      const filtered = []
      for (const emoji of this.list) {
        if (
          emoji.includes(this.search) || typeof this.keywords[emoji] !== 'undefined'
            ? this.keywords[emoji].some((keyword) => keyword.includes(this.search))
            : false
        ) {
          filtered.push(emoji)
        }
      }
      return filtered
    }

    scrollTo(event: MouseEvent, index: number) {
      if (!this._groups[index]) {
        return
      }
      this._scroll.scrollTop = index == 0 ? 0 : this._groups[index].offsetTop
    }

    onScroll() {
      if (!this.waitingForPaint) {
        this.waitingForPaint = true
        window.requestAnimationFrame(this.onScrollPaint.bind(this))
      }
    }

    onScrollPaint() {
      this.waitingForPaint = false
      let scrollTop = this._scroll.scrollTop
      let active = 0
      for (const [i] of this.groups.entries()) {
        let component = this._groups[i]
        if (component && component.offsetTop > scrollTop) {
          break
        }
        active = i
      }
      if (this.index !== active) {
        this.index = active
      }
    }

    onMouseExit() {
      this.hovered = ''
    }

    onMouseEnter(event: MouseEvent, emoji: string) {
      this.hovered = emoji
      this._search.placeholder = `:${emoji}:`
    }

    onClick(event: MouseEvent, emoji: string) {
      this.$accessor.emoji.setRecent(emoji)
      this.$emit('picked', emoji)
    }

    onClickAway() {
      this.$emit('done')
    }
  }
</script>
