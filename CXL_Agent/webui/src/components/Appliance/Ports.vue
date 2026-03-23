<!-- Copyright (c) 2024 Seagate Technology LLC and/or its Affiliates -->
<template>
  <v-container style="padding: 0">
    <v-data-table
      :headers="headers"
      fixed-header
      height="310"
      :items="portMemoryResourceRows"
    />
  </v-container>
</template>

<script>
import { computed } from "vue";
import { useBladePortStore } from "../Stores/BladePortStore";
import { useBladeMemoryStore } from "../Stores/BladeMemoryStore";
import { useBladeResourceStore } from "../Stores/BladeResourceStore";

export default {
  data() {
    return {
      headers: [
        { title: "PortId", align: "start", key: "portId" },
        { title: "LinkedHost/Port", key: "linkedHost" },
        { title: "LinkedMemoryId", key: "memoryId" },
        { title: "LinkedResourceId", key: "resourceId" },
        { title: "LinkedPPBId", key: "ppbId" },
        { title: "LinkedLDId", key: "ldId" },
      ],
    };
  },

  setup() {
    const bladePortStore = useBladePortStore();
    const bladeMemoryStore = useBladeMemoryStore();
    const bladeResourceStore = useBladeResourceStore();

    // (기존과 동일) Port 정렬 + NOT_FOUND 제외
    const sortedBladePorts = computed(() => {
      return bladePortStore.bladePorts
        .filter((port) => port.linkedPortUri != "NOT_FOUND")
        .slice()
        .sort((a, b) => {
          const numA = parseInt(a.id.replace(/^\D+/g, ""));
          const numB = parseInt(b.id.replace(/^\D+/g, ""));
          return numA - numB;
        });
    });

    // ✅ PortId - LinkedHost - LinkedMemory - ResourceId - PPBId 를 "완전히 펼친(flatten)" rows
    const portMemoryResourceRows = computed(() => {
      const ports = sortedBladePorts.value || [];
      const memories = bladeMemoryStore.bladeMemory || [];
      //const resources = bladeResourceStore.bladeResources || [];
      const resources = bladeResourceStore.memoryResources || [];

      // ResourceId -> PPBId(channelId) 맵
      const resourceMap = new Map(
        resources.map((r) => [
          r.id,
          {
            ppbId: r.channelId,
            ldId: r.channelResourceIndex
          },
        ])
      );

      const rows = [];

      for (const port of ports) {
        const portId = port.id;
        const linkedHost = port.linkedPortUri;

        // 이 포트에 연결된 모든 MemoryRegion
        const linkedMemories = memories.filter(
          (m) => m.memoryAppliancePort === portId
        );

        // ✅ (추가) 연결된 메모리가 0개인 포트도 1줄로 리스트업
        if (linkedMemories.length === 0) {
          rows.push({
            portId,
            linkedHost,
            memoryId: "-",
            resourceId: "-",
            ppbId: "-",
            ldId: "-",
          });
          continue;
        }

        // 메모리가 1개 이상이면: MemoryRegion 단위로 펼치고, 그 아래 resourceIds까지 펼침
        for (const mem of linkedMemories) {
          const memoryId = mem.id;

          const rids = mem.resourceIds || [];
          if (rids.length === 0) {
            rows.push({
              portId,
              linkedHost,
              memoryId,
              resourceId: "-",
              ppbId: "-",
              ldId: "-",
            });
            continue;
          }

          for (const resourceId of rids) {
            const map = resourceMap.get(resourceId) || {};

            rows.push({
              portId,
              linkedHost,
              memoryId,
              resourceId,
              ppbId: map.ppbId ?? "-",
              ldId: map.ldId ?? "-",
            });
          }
        }
      }

      return rows;
    });

    return {
      portMemoryResourceRows,
    };
  },
};
</script>
