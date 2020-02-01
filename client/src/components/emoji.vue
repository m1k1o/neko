<template>
  <div class="neko-emoji">
    <div class="search">
      <div class="search-contianer">
        <input type="text" ref="search" v-model="search" />
      </div>
    </div>
    <div class="list" ref="scroll" @scroll="onScroll">
      <ul :class="[search === '' ? 'group-list' : 'emoji-list']">
        <template v-if="search === ''">
          <li v-for="(group, index) in groups" :key="index" class="group" ref="groups">
            <span class="label">{{ group.name }}</span>
            <ul class="emoji-list">
              <li
                v-for="emoji in group.list"
                :key="`${group.id}-${emoji}`"
                :class="['emoji', hovered === emoji ? 'active' : '']"
              >
                <span
                  :class="['emoji-20', `e-${emoji}`]"
                  @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
                  @click.stop.prevent="onClick($event, emoji)"
                ></span>
              </li>
            </ul>
          </li>
        </template>
        <template v-else>
          <li v-for="emoji in filtered" :key="emoji" :class="['emoji', hovered === emoji ? 'active' : '']">
            <span
              :class="['emoji-20', `e-${emoji}`]"
              @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
              @click.stop.prevent="onClick($event, emoji)"
            ></span>
          </li>
        </template>
      </ul>
    </div>
    <div class="details">
      <div class="icon-container" v-if="hovered !== ''">
        <span :class="['icon', 'emoji-20', `e-${hovered}`]"></span><span class="emoji">:{{ hovered }}:</span>
      </div>
    </div>
    <div class="groups">
      <ul>
        <li
          v-for="(group, index) in groups"
          :key="index"
          :class="[group.id, active === group.id && search === '' ? 'active' : '']"
          @click.stop.prevent="scrollTo($event, index)"
        >
          <span :class="[`group-${group.id}`]" />
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
    bottom: 30px;
    right: 0;
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
          font-family: 'Font Awesome 5 Free';
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
          &.emoji {
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

      .icon-container {
        display: flex;
        align-content: center;
        flex-direction: row;
        height: 20px;

        span {
          cursor: default;

          &.icon {
            margin: 0 5px 0 10px;
          }

          &.emoji {
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

            &.group-recent {
              background-color: #fff;
            }
            &.group-neko {
              background-color: #fff;
            }
            &.group-people {
              background-color: #fff;
            }
            &.group-nature {
              background-color: #fff;
            }
            &.group-food {
              background-color: #fff;
            }
            &.group-activity {
              background-color: #fff;
            }
            &.group-travel {
              background-color: #fff;
            }
            &.group-objects {
              background-color: #fff;
            }
            &.group-symbols {
              background-color: #fff;
            }
            &.group-flags {
              background-color: #fff;
            }
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'

  import { list } from './emoji/list'
  import { keywords } from './emoji/keywords'
  import { groups } from './emoji/groups'

  @Component({
    name: 'neko-emoji',
  })
  export default class extends Vue {
    @Ref('scroll') readonly _scroll!: HTMLElement
    @Ref('search') readonly _search!: HTMLInputElement
    @Ref('groups') readonly _groups!: HTMLElement[]

    waitingForPaint = false
    search = ''
    active = groups[0].id
    hovered = ''

    get groups() {
      return groups
    }

    get filtered() {
      const filtered = []
      for (const emoji of list) {
        if (
          emoji.includes(this.search) || typeof keywords[emoji] !== 'undefined'
            ? keywords[emoji].some(keyword => keyword.includes(this.search))
            : false
        ) {
          filtered.push(emoji)
        }
      }
      return filtered
    }

    scrollTo(event: MouseEvent, index: number) {
      const ele = this._groups[index]
      if (!ele) {
        return
      }

      let top = ele.offsetTop
      if (index == 0) {
        top = 0
      }
      this._scroll.scrollTop = top
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
      let active = this.groups[0]
      for (let i = 0, l = this.groups.length; i < l; i++) {
        let group = this.groups[i]
        let component = this._groups[i]
        if (component && component.offsetTop > scrollTop) {
          break
        }
        active = group
      }
      this.active = active.id
    }

    onMouseExit(event: MouseEvent, emoji: string) {
      this.hovered = ''
    }

    onMouseEnter(event: MouseEvent, emoji: string) {
      this.hovered = emoji
      this._search.placeholder = `:${emoji}:`
    }

    onClick(event: MouseEvent, emoji: string) {
      this.$emit('picked', emoji)
    }
  }
</script>
