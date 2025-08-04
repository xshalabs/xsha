<?php

/**
 * 检查后端国际化翻译中未使用的键和已使用但未翻译的键
 * 扫描backend目录下的所有.go文件，查找翻译键使用情况
 */

class BackendI18nChecker {
    private $backendPath;
    private $i18nPath;
    private $srcPath;
    private $usedKeys = [];
    private $allKeys = [];
    private $missingKeys = [];
    
    public function __construct($basePath = '.') {
        $this->backendPath = rtrim($basePath, '/') . '/backend';
        $this->i18nPath = $this->backendPath . '/i18n/locales';
        $this->srcPath = $this->backendPath;
    }
    
    /**
     * 主执行方法
     */
    public function run() {
        echo "🔍 检查后端国际化翻译中未使用的键和已使用但未翻译的键...\n\n";
        
        // 检查路径是否存在
        if (!is_dir($this->i18nPath)) {
            die("❌ 错误: 翻译文件目录不存在: {$this->i18nPath}\n");
        }
        
        if (!is_dir($this->srcPath)) {
            die("❌ 错误: 后端源码目录不存在: {$this->srcPath}\n");
        }
        
        // 1. 读取所有翻译键
        $this->loadAllTranslationKeys();
        
        // 2. 扫描源码文件，查找使用的翻译键
        $this->scanSourceFiles();
        
        // 3. 找出未使用的键
        $this->findUnusedKeys();
        
        // 4. 找出已使用但未翻译的键
        $this->findMissingTranslations();
    }
    
    /**
     * 加载所有翻译键
     */
    private function loadAllTranslationKeys() {
        echo "📖 读取翻译文件...\n";
        
        // 使用 en-US.json 作为基准文件
        $baseFile = $this->i18nPath . '/en-US.json';
        
        if (!file_exists($baseFile)) {
            die("❌ 错误: 基准翻译文件不存在: {$baseFile}\n");
        }
        
        echo "   - 读取基准文件: en-US.json\n";
        
        $content = file_get_contents($baseFile);
        $data = json_decode($content, true);
        
        if ($data === null) {
            die("❌ 错误: 无法解析JSON文件 {$baseFile}\n");
        }
        
        $this->extractKeys($data);
        
        echo "✅ 总共找到 " . count($this->allKeys) . " 个翻译键\n\n";
    }
    
    /**
     * 递归提取所有翻译键
     */
    private function extractKeys($data, $prefix = '') {
        foreach ($data as $key => $value) {
            $fullKey = $prefix ? $prefix . '.' . $key : $key;
            
            if (is_array($value)) {
                $this->extractKeys($value, $fullKey);
            } else {
                // 使用键来避免重复
                $this->allKeys[$fullKey] = true;
            }
        }
    }
    
    /**
     * 扫描源码文件
     */
    private function scanSourceFiles() {
        echo "🔍 扫描Go源码文件中的翻译键使用...\n";
        
        $iterator = new RecursiveIteratorIterator(
            new RecursiveDirectoryIterator($this->srcPath)
        );
        
        $fileCount = 0;
        foreach ($iterator as $file) {
            if ($file->isFile() && preg_match('/\.go$/', $file->getFilename())) {
                $this->scanFile($file->getPathname());
                $fileCount++;
            }
        }
        
        echo "✅ 扫描了 {$fileCount} 个Go文件，找到 " . count($this->usedKeys) . " 个使用的翻译键\n\n";
    }
    
    /**
     * 扫描单个文件
     */
    private function scanFile($filePath) {
        $content = file_get_contents($filePath);
        
        // 匹配各种翻译使用模式
        $patterns = [
            // i18n.T(lang, "key") 或 i18n.T(lang, "key", args...)
            '/i18n\.T\s*\(\s*[^,]+,\s*["\']([^"\']+)["\']/m',
            // h.T("key") 或 h.T("key", args...)
            '/\.T\s*\(\s*["\']([^"\']+)["\']/m',
            // h.Response(c, statusCode, "messageKey", ...)
            '/\.Response\s*\(\s*[^,]+,\s*[^,]+,\s*["\']([^"\']+)["\']/m',
            // h.ErrorResponse(c, statusCode, "errorKey", ...)
            '/\.ErrorResponse\s*\(\s*[^,]+,\s*[^,]+,\s*["\']([^"\']+)["\']/m',
            // T(lang, "key", args...) - global T function
            '/\bT\s*\(\s*[^,]+,\s*["\']([^"\']+)["\']/m',
            // errors.New("translationKey") - error with translation key
            '/errors\.New\s*\(\s*["\']([a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z0-9_.]+)["\']\s*\)/m',
            // return errors.New("translationKey")
            '/return\s+errors\.New\s*\(\s*["\']([a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z0-9_.]+)["\']\s*\)/m'
        ];
        
        foreach ($patterns as $pattern) {
            preg_match_all($pattern, $content, $matches);
            
            if (!empty($matches[1])) {
                foreach ($matches[1] as $key) {
                    // 使用键来避免重复
                    $this->usedKeys[$key] = true;
                }
            }
        }
        
        // 特殊处理 error_mapping.go 文件
        if (basename($filePath) === 'error_mapping.go') {
            $this->scanErrorMappingFile($content);
        }
    }
    
    /**
     * 扫描 error_mapping.go 文件中的翻译键映射
     */
    private function scanErrorMappingFile($content) {
        // 匹配 ErrorMapping 中的键值对映射
        // 例如: "some error": "translation.key",
        $pattern = '/["\']([^"\']+)["\']\s*:\s*["\']([a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z0-9_.]+)["\']/m';
        
        preg_match_all($pattern, $content, $matches);
        
        if (!empty($matches[2])) {
            foreach ($matches[2] as $key) {
                // 使用键来避免重复
                $this->usedKeys[$key] = true;
            }
        }
    }
    
    /**
     * 查找未使用的键
     */
    private function findUnusedKeys() {
        echo "🔎 分析未使用的翻译键...\n\n";
        
        // 获取所有键和已使用键的数组
        $allKeysArray = array_keys($this->allKeys);
        $usedKeysArray = array_keys($this->usedKeys);
        
        // 检查已使用键中是否有不在总键列表中的
        $invalidUsedKeys = array_diff($usedKeysArray, $allKeysArray);
        if (!empty($invalidUsedKeys)) {
            echo "⚠️  发现 " . count($invalidUsedKeys) . " 个无效的已使用键（不在翻译文件中）:\n";
            foreach (array_slice($invalidUsedKeys, 0, 10) as $key) {
                echo "   - {$key}\n";
            }
            if (count($invalidUsedKeys) > 10) {
                echo "   - ... 还有 " . (count($invalidUsedKeys) - 10) . " 个\n";
            }
            echo "\n";
        }
        
        // 计算有效的已使用键和未使用键
        $validUsedKeys = array_intersect($usedKeysArray, $allKeysArray);
        $unusedKeys = array_diff($allKeysArray, $validUsedKeys);
        
        if (empty($unusedKeys)) {
            echo "🎉 太棒了！所有翻译键都被使用了！\n";
            return;
        }
        
        // 按模块分组显示未使用的键
        $groupedUnused = [];
        foreach ($unusedKeys as $key) {
            $parts = explode('.', $key);
            $module = $parts[0];
            $remainingKey = implode('.', array_slice($parts, 1));
            
            if (!isset($groupedUnused[$module])) {
                $groupedUnused[$module] = [];
            }
            $groupedUnused[$module][] = $remainingKey;
        }
        
        echo "❌ 发现 " . count($unusedKeys) . " 个未使用的翻译键：\n\n";
        
        ksort($groupedUnused); // 按模块名排序
        foreach ($groupedUnused as $module => $keys) {
            echo "📁 {$module} 模块 (" . count($keys) . " 个未使用):\n";
            sort($keys); // 按键名排序
            foreach ($keys as $key) {
                if ($key) {
                    echo "   - {$module}.{$key}\n";
                } else {
                    echo "   - {$module}\n";
                }
            }
            echo "\n";
        }
        
        // 统计信息
        echo "📊 统计信息:\n";
        echo "   - 总翻译键数: " . count($allKeysArray) . "\n";
        echo "   - 有效使用键数: " . count($validUsedKeys) . "\n";
        echo "   - 未使用键数: " . count($unusedKeys) . "\n";
        echo "   - 使用率: " . round((count($validUsedKeys) / count($allKeysArray)) * 100, 2) . "%\n\n";
        
        // 生成清理建议
        echo "💡 清理建议:\n";
        echo "   可以考虑删除这些未使用的翻译键以减少文件大小\n";
        echo "   删除前请确认这些键确实不会在动态生成的场景中使用\n";
        echo "   建议在删除前备份翻译文件\n\n";
        
        // 按使用频率显示最常见的模块
        $moduleUsage = [];
        foreach ($validUsedKeys as $key) {
            $parts = explode('.', $key);
            $module = $parts[0];
            if (!isset($moduleUsage[$module])) {
                $moduleUsage[$module] = 0;
            }
            $moduleUsage[$module]++;
        }
        
        arsort($moduleUsage);
        echo "📈 各模块使用情况 (按使用频率排序):\n";
        foreach ($moduleUsage as $module => $count) {
            $total = count(array_filter($allKeysArray, function($key) use ($module) {
                return strpos($key, $module . '.') === 0 || $key === $module;
            }));
            $usage = $total > 0 ? round(($count / $total) * 100, 1) : 0;
            echo "   - {$module}: {$count}/{$total} 个键被使用 ({$usage}%)\n";
        }
    }
    
    /**
     * 找出已使用但未翻译的键
     */
    private function findMissingTranslations() {
        echo "🔍 检查已使用但未翻译的键...\n\n";
        
        // 获取所有键和已使用键的数组
        $allKeysArray = array_keys($this->allKeys);
        $usedKeysArray = array_keys($this->usedKeys);
        
        // 找出已使用但不在翻译文件中的键
        $missingKeys = array_diff($usedKeysArray, $allKeysArray);
        
        if (empty($missingKeys)) {
            echo "✅ 太棒了！所有使用的翻译键都已定义！\n\n";
            return;
        }
        
        // 按模块分组显示缺失的键
        $groupedMissing = [];
        foreach ($missingKeys as $key) {
            $parts = explode('.', $key);
            $module = $parts[0];
            $remainingKey = implode('.', array_slice($parts, 1));
            
            if (!isset($groupedMissing[$module])) {
                $groupedMissing[$module] = [];
            }
            $groupedMissing[$module][] = $remainingKey;
        }
        
        echo "⚠️  发现 " . count($missingKeys) . " 个已使用但未翻译的键：\n\n";
        
        ksort($groupedMissing); // 按模块名排序
        foreach ($groupedMissing as $module => $keys) {
            echo "📁 {$module} 模块 (" . count($keys) . " 个缺失):\n";
            sort($keys); // 按键名排序
            foreach ($keys as $key) {
                if ($key) {
                    echo "   - {$module}.{$key}\n";
                } else {
                    echo "   - {$module}\n";
                }
            }
            echo "\n";
        }
        
        // 统计信息
        echo "📊 缺失键统计信息:\n";
        echo "   - 已使用键数: " . count($usedKeysArray) . "\n";
        echo "   - 缺失翻译键数: " . count($missingKeys) . "\n";
        echo "   - 翻译完整率: " . round(((count($usedKeysArray) - count($missingKeys)) / count($usedKeysArray)) * 100, 2) . "%\n\n";
        
        // 生成修复建议
        echo "💡 修复建议:\n";
        echo "   需要在翻译文件中添加这些缺失的键\n";
        echo "   建议统一添加到 backend/i18n/locales/en-US.json 和 zh-CN.json 中\n";
        echo "   可以先添加占位符文本，后续再完善翻译内容\n\n";
        
        // 生成JSON格式的缺失键，方便复制添加
        echo "📋 建议添加到翻译文件的JSON格式:\n";
        echo "```json\n";
        sort($missingKeys);
        foreach ($missingKeys as $key) {
            echo "  \"{$key}\": \"[需要翻译] {$key}\",\n";
        }
        echo "```\n\n";
    }
}

// 执行检查
try {
    $checker = new BackendI18nChecker();
    $checker->run();
} catch (Exception $e) {
    echo "❌ 执行出错: " . $e->getMessage() . "\n";
    exit(1);
}

?>