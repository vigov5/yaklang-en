package yso

const (
	// CommonsCollections1/3/5/6/7 chains, the required<=3.2.1 version
	CC31Or321 = "org.apache.commons.collections.functors.ChainedTransformer"
	CC322     = "org.apache.commons.collections.ExtendedProperties$1"
	CC40      = "org.apache.commons.collections4.functors.ChainedTransformer"
	CC41      = "org.apache.commons.collections4.FluentIterable"
	// CommonsBeanutils2 chain, serialVer sionUID is different, 1.7x-1.8 x is -3490850999041592962, 1.9x is -2044202215314119608
	CB17  = "org.apache.commons.beanutils.MappedPropertyDescriptor$1"
	CB18x = "org.apache.commons.beanutils.DynaBeanMapDecorator$MapEntry"
	CB19x = "org.apache.commons.beanutils.BeanIntrospectionData"
	//c3p0 serialVersionUID is different, 0.9.2pre2-0.9.5pre8 is 7387108436934414104, 0.9.5pre9-0.9.5.5 is 7387108436934414104
	C3p092x = "com.mchange.v2.c3p0.impl.PoolBackedDataSourceBase"
	C3p095x = "com.mchange.v2.c3p0.test.AlwaysFailDataSource"
	// AspectJWeaver requires cc31
	Ajw = "org.aspectj.weaver.tools.cache.SimpleCache"
	// bsh serialVersionUID is different, 2.0b4 is 4949939576606791809, 2.0b5 is 4041428789013517368, 2.0.b6 cannot deserialize
	Bsh20b4 = "bsh.CollectionManager$1"
	Bsh20b5 = "bsh.engine.BshScriptEngine"
	Bsh20b6 = "bsh.collection.CollectionIterator$1"
	// Groovy 1.7.0-2.4.3 based on Jdk7u21. The serialVersionUID is different. 2.4.x is -8137949907733646644, and 2.3.x is 1228988487386910280. There is a chain
	Groovy1702311 = "org.codehaus.groovy.reflection.ClassInfo$ClassInfoSet"
	Groovy24x     = "groovy.lang.Tuple2"
	Groovy244     = "org.codehaus.groovy.runtime.dgm$1170"
	// Becl JDK<8u251
	Becl                = "com.sun.org.apache.bcel.internal.util.ClassLoader"
	DefiningClassLoader = "org.mozilla.javascript.DefiningClassLoader"
	Jdk7u21             = "com.sun.corba.se.impl.orbutil.ORBClassLoader"
	// JRE8u20 7u25<=JDK<=8u20, although it is called JRE8u20, JDK8u20 can also be used, this detection is not perfect, 8u25 version and JDK<=7u21, which will cause false positives. You can look at
	JRE8u20 = "javax.swing.plaf.metal.MetalFileChooserUI$DirectoryComboBoxModel$1"
	// ROME1000 Rome <= 1.11.1
	ROME1000 = "com.sun.syndication.feed.impl.ToStringBean"
	ROME1111 = "com.rometools.rome.feed.impl.ObjectBean"
	// Fastjson fastjson<=1.2.48. , all versions need to use hashMap to bypass checkAutoType
	// This chain relies on BadAttributeValueExpException and cannot be used in JDK1.7. At this time, springAOP needs to be used to bypass
	Fastjson = "com.alibaba.fastjson.JSONArray"
	// Jackson jackson-databind>=2.10 .0 There is a chain
	Jackson = "com.fasterxml.jackson.databind.node.NodeSerialization"
	// SpringAOP fastjon/jackson variants of both chains require springAOP
	SpringAOP = "org.springframework.aop.target.HotSwappableTargetSource.HotSwappableTargetSource"
	LinuxOS   = "sun.awt.X11.AwtGraphicsConfigData"
	WindowsOS = "sun.awt.windows.WButtonPeer"
)

var allGadgetsCheckList = map[string]string{
	"CC31Or321":           CC31Or321,
	"CC322":               CC322,
	"CC40":                CC40,
	"CC41":                CC41,
	"CB17":                CB17,
	"CB18x":               CB18x,
	"CB19x":               CB19x,
	"C3p092x":             C3p092x,
	"C3p095x":             C3p095x,
	"Ajw":                 Ajw,
	"Bsh20b4":             Bsh20b4,
	"Bsh20b5":             Bsh20b5,
	"Bsh20b6":             Bsh20b6,
	"Groovy1702311":       Groovy1702311,
	"Groovy24x":           Groovy24x,
	"Groovy244":           Groovy244,
	"Becl":                Becl,
	"DefiningClassLoader": DefiningClassLoader,
	"Jdk7u21":             Jdk7u21,
	"JRE8u20":             JRE8u20,
	"ROME1000":            ROME1000,
	"ROME1111":            ROME1111,
	"Fastjson":            Fastjson,
	"Jackson":             Jackson,
	"SpringAOP":           SpringAOP,
	"Linux_OS":            LinuxOS,
	"Windows_OS":          WindowsOS,
}

func GetGadgetChecklist() map[string]string {
	return allGadgetsCheckList
}
