# Source Baseline（冻结，实施不得回写）

冻结日期2026-11-21；源码范围仅v0.2.23..v0.2.24。真实实施起点填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `d2488b1105f091b6fce4445395fec1835306fcf8` / `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| VERSION / Go / SQLite / auth snapshot | `1.1.26 / 1.27.4 / 230 / 21` |
| 私有 index tree | `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| v0.2.24 commit / tag object | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` / `7515db6dc56c2a6120dfb7abed70ef089ebc1457` |
| upstream/main（观察） | `328db26f450edbdf9f8d4344a3ea903ac6396842` |
| 非merge / 完整候选 | 65 / 39 PR + 1普通直接提交 + 1无PR合并簇 |
| 净diff / 候选×文件态 | 190文件 +11853 -610 / 217条 |
| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 135 / 62 / 20 |

| 候选 | merge/direct | first parent | 完整patch SHA256 |
|---|---|---|---|
| [#8671](https://github.com/Wei-Shaw/sub2api/pull/8671) | `c38a534be9a2e9600084b5f514badad828886958` | `16b53b60a8c9b83a6e4e93c8157b9b7756dc2c30` | `e551ef37d90f0f1729d9a92fb6a20a8600e400e69fa558b765c14eb2d1c3535d` |
| [#8820](https://github.com/Wei-Shaw/sub2api/pull/8820) | `600e1e1888badd20e2adb4a14da15fe8605391ef` | `135aff88e12db62446a7f454b01b864cbbfc13c6` | `3808a870aeca0ca9f3647b65dba3863d448efdf16a97e55376a83e69cff19591` |
| [#8845](https://github.com/Wei-Shaw/sub2api/pull/8845) | `84a293056a000715b38f66ad41b6f8b2484fc32e` | `600e1e1888badd20e2adb4a14da15fe8605391ef` | `6d2e36e33ae9f5bb05ae6aa150e6d277af969e3f0cfe9f5baf7234114f55a75c` |
| [#8858](https://github.com/Wei-Shaw/sub2api/pull/8858) | `29d9fe2ad0e091c104a24e916ff708ec057570cf` | `84a293056a000715b38f66ad41b6f8b2484fc32e` | `807f3df84a3ca7d8c38b48c72157e44d7b1843068d27010ae05f8b5ed02acdb4` |
| [#8859](https://github.com/Wei-Shaw/sub2api/pull/8859) | `3ad84e1333f4ae05ac1ce20221a940f2f2095e3d` | `29d9fe2ad0e091c104a24e916ff708ec057570cf` | `48092f37553f4a813d70fbde936c6a1f1b66c9779426d854411fb5fd08733279` |
| [#8862](https://github.com/Wei-Shaw/sub2api/pull/8862) | `998d1c8e2ee2c59bc0e8b60b3a5dc62e4e20f148` | `3ad84e1333f4ae05ac1ce20221a940f2f2095e3d` | `ee96c65ca8e3b45029c36962fdc7c11371d01539cc56ef52083abbd560dca45f` |
| [#8864](https://github.com/Wei-Shaw/sub2api/pull/8864) | `9cf9921eb4e5b61b8eeb8139f6d8099fe3325725` | `998d1c8e2ee2c59bc0e8b60b3a5dc62e4e20f148` | `a983d0dd2e0d9c9c9694fb000af184a8172dcbcdb2b3d31972b57feb308bb8c8` |
| [#8866](https://github.com/Wei-Shaw/sub2api/pull/8866) | `50c802c518de49bdd259637b9f71df3b8e2f9917` | `9cf9921eb4e5b61b8eeb8139f6d8099fe3325725` | `51d0986c3d122fe451549ecb30d0113cc59cd55cc12b0f20beb4208d83e16321` |
| [#8868](https://github.com/Wei-Shaw/sub2api/pull/8868) | `a437224cd291f5cfa48d95252919e650d38c7141` | `50c802c518de49bdd259637b9f71df3b8e2f9917` | `7f0f86d44a0c5bd3fbf6782dd7c7245e67d71030093d4ed226adc62d7e605cf9` |
| [#8870](https://github.com/Wei-Shaw/sub2api/pull/8870) | `67b2df8d91ffc93c13fa176810008d14376c7e7b` | `a437224cd291f5cfa48d95252919e650d38c7141` | `6748e6c9c924b8d110d36515fce3639dc1e7234af758ea406781933bcadc4cdb` |
| [#8872](https://github.com/Wei-Shaw/sub2api/pull/8872) | `356cdccbe888854cf631127c9754b6051fa17e07` | `67b2df8d91ffc93c13fa176810008d14376c7e7b` | `36929a0cc140b1be61de7d22d13d5a9ab10abfeb58e115d9dc1f12dff4d3c927` |
| [#8875](https://github.com/Wei-Shaw/sub2api/pull/8875) | `fb7d8fc7b91a101fe69800c3dca3917c83c32eaf` | `356cdccbe888854cf631127c9754b6051fa17e07` | `075327d39bff83e02a6367c617a714f603a1d87ac3c5805d5b04fc1ef19c317d` |
| [#8883](https://github.com/Wei-Shaw/sub2api/pull/8883) | `490af330d7a7c3e7084c9ad6f50e5ba7766527a9` | `fb7d8fc7b91a101fe69800c3dca3917c83c32eaf` | `1e0a6507941860b555d4ec1583842577d26324f2d7dea41f5a3972559524de70` |
| [#8893](https://github.com/Wei-Shaw/sub2api/pull/8893) | `fe2d953f528ae05fd6fca494431a096fa6272d7e` | `490af330d7a7c3e7084c9ad6f50e5ba7766527a9` | `4d4dab43b9ee93f5166d4a8069d6067e85bdec6f5cd5a86ea4eca759e9c7600a` |
| [#8899](https://github.com/Wei-Shaw/sub2api/pull/8899) | `6eb32ab4451ef51f6dea039bd840175d19d47e56` | `fe2d953f528ae05fd6fca494431a096fa6272d7e` | `8377b233a2b10522112c5490171bb29ad91c453eb3949c9777f4cbe1d7e16258` |
| [#8918](https://github.com/Wei-Shaw/sub2api/pull/8918) | `a912d13e62514cd4d6233126da0cd77bef901d2b` | `69237d2ebd166674739b8ee5e6a343d0f3f635b4` | `9b10d42b537cf0a4440244c211c34ff67aa33bebf73efa6a43ff62a96346e80b` |
| [#8932](https://github.com/Wei-Shaw/sub2api/pull/8932) | `320695ae72639efba68e25e1ff5094f7fbea46b9` | `fe1f2b6d76cc1c535ae7bcebdc0e694aac9cb742` | `df59e409ae453d6ca964f892f878ab976c59f87a75478ffae0ee5973f8de84e8` |
| [#8937](https://github.com/Wei-Shaw/sub2api/pull/8937) | `1134f5dffbf96b4c9a82d81e2ec4f93c5cabbf06` | `320695ae72639efba68e25e1ff5094f7fbea46b9` | `1da87152a7ce400670864438864d1890205a8e53171aa7f5ce31774d5fe655cd` |
| [#8973](https://github.com/Wei-Shaw/sub2api/pull/8973) | `242556f973a27d8bc86af769fe814a983825d55a` | `cdf3af4b84f884d78928155d126848259c932bb4` | `8970ab99a7fea039311222d6a3ef42deaa8d3fabf438f447ed8eaece3b4dab85` |
| [#8979](https://github.com/Wei-Shaw/sub2api/pull/8979) | `eb9143956f1e74458c6ca56af069c3e8604c3bd0` | `1eea0d750a945c7c83a44b7bb8ff6232fc03521e` | `c6590343e397ec2e3e2f95a775faba1df8b32bce03e619520ff59870433abe58` |
| [5c2982a81](https://github.com/Wei-Shaw/sub2api/commit/5c2982a81b46e2142d0143d2f83c1ce9c4da1d24) | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` | `62cebadf54b5b208d6dd9b93f1929d25263e67e8` | `8d05ba92a2bc9840d70843686cf0e62fa283c6540d4b9f0d6378520403bbaf7d` |

最终tag优先于中间提交。全部候选先reverse再forward；来源完整性与评估证据见[MANIFEST](../../../docs/upstream-sync/evidence-0.2.24/MANIFEST.md)。不把来源的PG/缺平台接口当成本地实现。
